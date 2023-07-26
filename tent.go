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

func Tent(bs BatchSystem, cmdline []string, mergeOutput bool) (err error) {
	tentVariables := bs.TentVariables()

	go func() {
		for retry := 0; retry < maxRetries; retry++ {
			if err := tent.RunShellServer(); err != nil {
				log.Printf("spanner: could not start shell server: %v", err)
			}
			time.Sleep(retryDelay)
		}
	}()

	process, err := tent.RunCommand(cmdline, tentVariables, mergeOutput)
	if err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	_, err = process.Wait()
	if err != nil {
		return fmt.Errorf("waiting on the process: %w", err)
	}
	return nil
}
