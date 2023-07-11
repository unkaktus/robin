package pbs

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/unkaktus/spanner"
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

func (b *PBS) ClearHistory() error {
	jobList, err := b.ListJobs(false)
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

func (b *PBS) clearInvisibleJobs(jobList []spanner.Job) error {
	jobMap := map[string]spanner.Job{}
	for _, job := range jobList {
		addedJob, ok := jobMap[job.Name]
		if ok {
			if job.CreationTime.After(addedJob.CreationTime) {
				if err := wipeJob(addedJob.ID); err != nil {
					return fmt.Errorf("wipe invisible job: %w", err)
				}
				jobMap[job.Name] = job
			}
		} else {
			jobMap[job.Name] = job
		}
	}
	return nil
}
