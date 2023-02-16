package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func cancel(jobName string) error {
	jobList, err := listJobs()
	if err != nil {
		return fmt.Errorf("list jobs: %w", err)
	}
	found := false
	for _, job := range jobList {
		if job.Name == jobName {
			found = true
			cmd := exec.Command("qdel", "-x", job.ID)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("execute qdel: %w", err)
			}
			log.Printf("cancelled %s (%s)", job.Name, job.ID)
		}
	}
	if !found {
		return fmt.Errorf("job not found")
	}
	return nil
}
