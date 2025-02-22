package tmux

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func (b *Tmux) Cancel(jobName string) error {
	jobList, err := b.ListJobs(false)
	if err != nil {
		return fmt.Errorf("list jobs: %w", err)
	}
	found := false
	for _, job := range jobList {
		if job.Name == jobName {
			found = true
			cmd := exec.Command("tmux", "kill-session", "-t", job.ID)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("execute qsig: %w", err)
			}
			log.Printf("cancelled %s", job.Name)
		}
	}
	if !found {
		return fmt.Errorf("job not found")
	}
	return nil

}
