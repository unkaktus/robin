package spanner

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func On(machine string, cmdline []string) error {
	cmd := exec.Command("ssh", []string{
		"-tt",
		"-q",
		machine,
		fmt.Sprintf("$SHELL -l -c 'spanner %s'", strings.Join(cmdline, " ")),
	}...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute on remote: %w", err)
	}
	return nil
}
