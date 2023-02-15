package main

import (
	"fmt"
	"os"
	"os/exec"
)

func logs(jobName string, outputType string) error {
	jobList, err := listJobs()
	if err != nil {
		return fmt.Errorf("list jobs: %w", err)
	}
	for _, job := range jobList {
		if job.Name == jobName {
			logFile := job.OutputFile
			if outputType == "err" {
				logFile = job.ErrorFile
			}
			cmd := exec.Command("tail", "-n", "128", "-f", logFile)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("execute tail: %w", err)
			}
			break
		}
	}
	return nil
}
