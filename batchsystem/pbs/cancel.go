package pbs

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

var (
	cmdTerminate = []string{"qsig", "-s", "SIGTERM"}
	cmdDelete    = []string{"qdel", "-x"}
)

func (b *PBS) Cancel(jobName string) error {
	jobList, err := b.ListJobs(true)
	if err != nil {
		return fmt.Errorf("list jobs: %w", err)
	}
	found := false
	for _, job := range jobList {
		if job.Name == jobName {
			found = true
			var cmdline []string
			if job.State == "R" {
				cmdline = cmdTerminate
			} else {
				cmdline = cmdDelete
			}
			cmd := exec.Command(cmdline[0], append(cmdline[1:], job.ID)...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("execute qsig: %w", err)
			}
			log.Printf("cancelled %s (%s)", job.Name, job.ID)
		}
	}
	if !found {
		return fmt.Errorf("job not found")
	}
	return nil
}
