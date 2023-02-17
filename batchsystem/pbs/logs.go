package pbs

import (
	"fmt"
	"os"
	"os/exec"
)

func (b *PBS) Logs(jobName string, outputType string) error {
	jobList, err := b.ListJobs()
	if err != nil {
		return fmt.Errorf("list jobs: %w", err)
	}
	for _, job := range jobList {
		if job.Name == jobName {
			logFile := job.OutputFile
			if outputType == "err" {
				logFile = job.ErrorFile
			}
			cmd := exec.Command("vim", logFile)
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

func (b *PBS) Logtail(jobName string, outputType string) error {
	jobList, err := b.ListJobs()
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
