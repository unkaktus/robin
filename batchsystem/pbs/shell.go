package pbs

import (
	"fmt"

	"github.com/unkaktus/spanner"
)

func (b *PBS) Shell(jobName string, nodeID int) error {
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

			if err := spanner.Shell(node); err != nil {
				return fmt.Errorf("execute ssh: %w", err)
			}
			break
		}
	}
	return nil
}
