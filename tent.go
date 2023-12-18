package spanner

import (
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/unkaktus/spanner/tent"
)

const (
	maxRetries int           = 5
	retryDelay time.Duration = 1 * time.Second
)

func Tent(bs BatchSystem, cmdline []string, noCommand bool) error {
	tentVariables := bs.TentVariables()
	nodeHead := make(chan struct{})

	go func() {
		var err error
		for retry := 0; retry < maxRetries; retry++ {
			if err = tent.RunShellServer(nodeHead); err != nil {
				time.Sleep(retryDelay)
				continue
			}
			break
		}
		if err != nil {
			log.Err(err).Msg("spanner: could not start shell server")
		}
	}()

	go func() {
		<-nodeHead
		cmd := exec.Command("node_exporter")
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		if err := cmd.Run(); err != nil {
			log.Err(err).Msg("run node head")
		}
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
