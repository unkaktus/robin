package tmux

import (
	"fmt"
	"os"

	"github.com/unkaktus/robin"
	"github.com/unkaktus/robin/shell"
)

func (b *Tmux) Shell(target *shell.Target, command string, verbose bool) error {
	job, err := b.FindJob(target.JobName)
	if err != nil {
		return fmt.Errorf("find job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found")
	}

	if target.NodeID < 0 || target.NodeID >= len(job.Nodes) {
		return fmt.Errorf("node ID is outside the node list range")
	}
	node := job.Nodes[target.NodeID]

	for {
		err := robin.Shell(node, command, shell.TargetPrompt(target))
		if err == nil {
			break
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "robin shell error: %v\n", err)
		}
	}

	return nil
}
