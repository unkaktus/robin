package pbs

import (
	"fmt"
	"os"

	"github.com/unkaktus/robin"
)

func (b *PBS) Shell(jobName string, nodeID int, verbose bool) error {
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

			for {
				err := robin.Shell(node)
				if err == nil {
					break
				}
				if verbose {
					fmt.Fprintf(os.Stderr, "robin shell error: %v\n", err)
				}
			}
		}
	}
	return nil
}
