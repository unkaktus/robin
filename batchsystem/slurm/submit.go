package slurm

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func (b *Slurm) Submit(jobData string) error {
	jobDataReader := strings.NewReader(jobData)
	cmd := exec.Command("sbatch")
	cmd.Stdin = jobDataReader
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute sbatch: %w", err)
	}
	return nil
}
