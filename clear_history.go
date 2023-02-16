package main

import (
	"fmt"
	"os"
	"os/exec"
)

func wipeJob(jobID string) error {
	cmd := exec.Command("qdel", "-x", jobID)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("qdel: %w", err)
	}
	return nil
}

func clearHistory() error {
	jobList, err := listJobs()
	if err != nil {
		return fmt.Errorf("query  job list: %w", err)
	}

	for _, job := range jobList {
		if job.State != "F" {
			continue
		}
		if err := wipeJob(job.ID); err != nil {
			return fmt.Errorf("removing %s: %w", job.ID, err)
		}
	}

	return nil
}
