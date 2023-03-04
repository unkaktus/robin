package spanner

import (
	"fmt"

	"github.com/unkaktus/spanner/tent"
)

func Tent(bs BatchSystem, cmdline []string) (err error) {
	tentVariables := bs.TentVariables()

	process, err := tent.RunCommand(cmdline, tentVariables)
	if err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	_, err = process.Wait()
	if err != nil {
		return fmt.Errorf("waiting on the process: %w", err)
	}
	return nil
}
