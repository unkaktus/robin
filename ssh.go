package main

import (
	"fmt"
	"os"
	"os/exec"
)

func ssh(jobName string, nodeID int) error {
	jobList, err := listJobs()
	if err != nil {
		return fmt.Errorf("list jobs: %w", err)
	}
	for _, job := range jobList {
		if job.Name == jobName {
			if nodeID < 0 || nodeID >= len(job.Nodes) {
				return fmt.Errorf("node ID is outside the node list range")
			}
			node := job.Nodes[nodeID]
			cmd := exec.Command("/usr/bin/env", "PBS_JOBID="+job.ID, "ssh", node)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("execute ssh: %w", err)
			}
			break
		}
	}
	return nil
}
