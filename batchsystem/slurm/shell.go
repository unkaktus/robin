package slurm

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func isSupermuc() bool {
	cmd := exec.Command("hostname", "-d")
	combi, _ := cmd.CombinedOutput()
	return strings.Contains(string(combi), "sng.lrz.de")
}

func (b *Slurm) Shell(jobName string, nodeID int, nodeSuffux string) error {
	// XXX: in case it is SuperMUC, set opa route
	if isSupermuc() {
		nodeSuffux = "opa"
	}

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
			cmd := exec.Command("ssh",
				[]string{
					"-p", "2222",
					"-o", "LogLevel=ERROR",
					"-o", "UserKnownHostsFile=/dev/null",
					"-o", "StrictHostKeyChecking=no",
					node + nodeSuffux}...)
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
