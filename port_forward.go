package robin

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func findJobOnRemote(machine, jobName string) (*Job, error) {
	cmd := exec.Command("ssh", []string{
		"-tt",
		"-q",
		machine,
		"$SHELL -l -c 'echo robin; robin list --json'",
	}...)
	cmd.Stderr = os.Stderr
	stdout := &strings.Builder{}
	cmd.Stdout = stdout
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("execute on remote: %w", err)
	}

	// Filter login stuff out
	output := strings.SplitAfterN(stdout.String(), "robin", 2)
	if len(output) != 2 {
		log.Printf("data: %+v", output)
		return nil, fmt.Errorf("wrong output length: %v", len(output))
	}

	jobList := []Job{}
	err = json.Unmarshal([]byte(output[1]), &jobList)
	if err != nil {
		return nil, fmt.Errorf("unmarshal JSON output: %w", err)
	}
	for _, job := range jobList {
		if job.Name == jobName {
			return &job, nil
		}
	}
	return nil, fmt.Errorf("job not found")
}

func PortForward(machine, jobName string, port, nodeID int) error {
	job, err := findJobOnRemote(machine, jobName)
	if err != nil {
		return fmt.Errorf("list nodes on remote: %w", err)
	}
	if nodeID < 0 || nodeID >= len(job.Nodes) {
		return fmt.Errorf("node ID is outside the node list range")
	}
	node := job.Nodes[nodeID]

	log.Printf("Forwarding port %d to %s:%d", port, node, port)

	cmd := exec.Command("ssh", []string{
		"-N",
		"-q",
		fmt.Sprintf("-L %d:%s:%d", port, node, port),
		machine,
	}...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute ssh with port-forward: %w", err)
	}
	return nil
}
