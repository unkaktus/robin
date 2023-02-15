package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/olekukonko/tablewriter"
)

type ListOutput struct {
	Jobs map[string]struct {
		Name          string `json:"Job_Name"`
		State         string `json:"job_state"`
		Queue         string `json:"queue"`
		ExecHosts     string `json:"exec_host"`
		ResourcesUsed struct {
			CPUTime string `json:"cput"`
		} `json:"resources_used"`
	} `json:"Jobs"`
}

func query() (*ListOutput, error) {
	cmd := exec.Command("qstat", "-f", "-F", "json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("exectute command: %w", err)
	}

	listOutput := &ListOutput{}

	err = json.Unmarshal(out, listOutput)
	if err != nil {
		return nil, fmt.Errorf("unmarshal JSON: %w", err)
	}
	return listOutput, nil
}

func showTable(listOutput *ListOutput) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "ID", "State", "Queue", "Time", "Nodes"})

	for jobID, job := range listOutput.Jobs {
		table.Append([]string{
			job.Name,
			jobID,
			job.State,
			job.Queue,
			job.ResourcesUsed.CPUTime,
			job.ExecHosts,
		})
	}
	table.Render()
	return nil
}

func list() error {
	listOutput, err := query()
	if err != nil {
		return fmt.Errorf("query list: %w", err)
	}

	if err := showTable(listOutput); err != nil {
		return fmt.Errorf("query list: %w", err)
	}

	return nil
}
