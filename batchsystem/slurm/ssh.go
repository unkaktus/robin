package slurm

import (
	"fmt"
	"os"
	"os/exec"
)

func (b *Slurm) SSH(jobName string, nodeID int) error {
	jobList, err := b.ListJobs(true)
	if err != nil {
		return fmt.Errorf("list jobs: %w", err)
	}
	found := false
	for _, job := range jobList {
		if job.Name == jobName {
			found = true
			if nodeID < 0 || nodeID >= len(job.Nodes) {
				return fmt.Errorf("node ID is outside the node list range")
			}
			node := job.Nodes[nodeID]
			cmd := exec.Command("ssh", node)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("execute ssh: %w", err)
			}
			break
		}
	}
	if !found {
		return fmt.Errorf("job not found")
	}
	return nil
}
