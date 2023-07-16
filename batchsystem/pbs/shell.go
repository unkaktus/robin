package pbs

import (
	"fmt"
	"os"
	"os/exec"
)

func (b *PBS) Shell(jobName string, nodeID int, nodeSuffix string) error {
	jobList, err := b.ListJobs(true)
	if err != nil {
		return fmt.Errorf("list jobs: %w", err)
	}
	for _, job := range jobList {
		if job.Name == jobName {
			if nodeID < 0 || nodeID >= len(job.Nodes) {
				return fmt.Errorf("node ID is outside the node list range")
			}
			node := job.Nodes[nodeID]
			cmd := exec.Command("ssh",
				[]string{
					"-p", "2222",
					"-o", "LogLevel=ERROR",
					"-o", "UserKnownHostsFile=/dev/null",
					"-o", "StrictHostKeyChecking=no",
					node + nodeSuffix}...)
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
