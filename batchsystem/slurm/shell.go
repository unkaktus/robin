package slurm

import (
	"fmt"

	"github.com/unkaktus/robin"
)

func (b *Slurm) Shell(jobName string, nodeID int) error {
	// In case it is SuperMUC, set opa route

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

			node = robin.RewriteNode(node)

			if err := robin.Shell(node); err != nil {
				return fmt.Errorf("connect via ssh: %w", err)
			}
			break
		}
	}
	if !found {
		return fmt.Errorf("job not found")
	}
	return nil
}
