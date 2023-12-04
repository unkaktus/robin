package spanner

import (
	"fmt"
	"log"
	"time"

	"github.com/unkaktus/spanner/tent"
)

const (
	maxRetries int           = 5
	retryDelay time.Duration = 1 * time.Second
)

func Tent(bs BatchSystem, cmdline []string, noCommand bool) error {
	tentVariables := bs.TentVariables()

	go func() {
		var err error
		for retry := 0; retry < maxRetries; retry++ {
			if err = tent.RunShellServer(); err != nil {
				time.Sleep(retryDelay)
				continue
			}
			break
		}
		log.Printf("spanner: could not start shell server: %v", err)
	}()

	if noCommand {
		select {}
	} else {
		process, err := tent.RunCommand(cmdline, tentVariables)
		if err != nil {
			return fmt.Errorf("running command: %w", err)
		}
		err = process.Wait()
		if err != nil {
			return fmt.Errorf("waiting on the process: %w", err)
		}
	}
	return nil
}
