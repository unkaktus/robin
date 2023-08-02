package spanner

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func Logs(b BatchSystem, jobName string, outputType string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	jobList, err := b.ListJobs(true)
	if err != nil {
		return fmt.Errorf("list jobs: %w", err)
	}
	for _, job := range jobList {
		if job.Name == jobName {
			logFile := job.OutputFile
			if outputType == "err" {
				logFile = job.ErrorFile
			}
			cmd := exec.Command(editor, logFile)
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

func Logtail(b BatchSystem, jobName, outputType string, nLines int) error {
	jobList, err := b.ListJobs(true)
	if err != nil {
		return fmt.Errorf("list jobs: %w", err)
	}
	for _, job := range jobList {
		if job.Name == jobName {
			logFile := job.OutputFile
			if outputType == "err" {
				logFile = job.ErrorFile
			}
			cmd := exec.Command("tail", "-n", strconv.Itoa(nLines), "-f", logFile)
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
