package robin

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	robinJobFilename = "robin.job"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func Submit(bs BatchSystem, name string) error {
	var jobData []byte
	var err error
	// Check whether it's a plain job file
	if fileExists(name) {
		jobData, err = os.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read job data file: %w", err)
		}
	} else {
		// Ask robin.job executable to generate the job data
		cmd := exec.Command("./"+robinJobFilename, name)
		jobData, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("exectute %s command: %w", robinJobFilename, err)
		}
	}

	if err := bs.Submit(string(jobData)); err != nil {
		return fmt.Errorf("submit: %w", err)
	}
	return nil
}

func Resubmit(bs BatchSystem, name string) error {
	// Extract the job data filename from the comment
	job, err := findJob(bs, name)
	if err != nil {
		return fmt.Errorf("find job: %w", err)
	}

	comment := &Comment{}
	if err := comment.Decode(job.Comment); err != nil {
		return fmt.Errorf("decode comment: %w", err)
	}

	if comment.JobDataFilename == "" {
		return fmt.Errorf("no job data file")
	}

	jobData, err := os.ReadFile(comment.JobDataFilename)
	if err != nil {
		return fmt.Errorf("read job data file: %w", err)
	}

	// Cancel the job
	if err := bs.Cancel(name); err != nil {
		return fmt.Errorf("cancel error: %w", err)
	}

	// Submit with the same job data
	if err := bs.Submit(string(jobData)); err != nil {
		return fmt.Errorf("submit: %w", err)
	}
	return nil
}
