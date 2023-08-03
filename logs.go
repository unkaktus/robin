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

	job, err := findJob(b, jobName)
	if err != nil {
		return fmt.Errorf("find job: %w", err)
	}

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

	return nil
}

func Logtail(b BatchSystem, jobName, outputType string, nLines int) error {
	job, err := findJob(b, jobName)
	if err != nil {
		return fmt.Errorf("find job: %w", err)
	}

	logFile := job.OutputFile
	if outputType == "err" {
		logFile = job.ErrorFile
	}
	cmd := exec.Command("tail", "-n", strconv.Itoa(nLines), "-F", logFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute tail: %w", err)
	}
	return nil
}
