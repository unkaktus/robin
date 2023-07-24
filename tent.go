package spanner

import (
	"fmt"

	"github.com/unkaktus/spanner/tent"
)

func Tent(bs BatchSystem, cmdline []string, mergeOutput bool) (err error) {
	tentVariables := bs.TentVariables()

	go tent.RunShellServer()

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
