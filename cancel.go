package main

import (
	"fmt"
	"os"
	"os/exec"
)

func cancel(jobName string) error {
	jobList, err := listJobs()
	if err != nil {
		return fmt.Errorf("list jobs: %w", err)
	}
	for _, job := range jobList {
		if job.Name == jobName {
			cmd := exec.Command("qdel", job.ID)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("execute qdel: %w", err)
			}
			break
		}
	}
	return nil
}
