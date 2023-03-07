package pbs

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func (b *PBS) Submit(jobData string) error {
	jobDataReader := strings.NewReader(jobData)
	cmd := exec.Command("qsub")
	cmd.Stdin = jobDataReader
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute qsub: %w", err)
	}
	return nil
}
