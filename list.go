package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func showTable(jobList []Job) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "ID", "State", "Queue", "Time", "Nodes"})

	for _, job := range jobList {
		timePercentage := int(100 * job.Walltime.Seconds() / job.RequestedWalltime.Seconds())
		table.Append([]string{
			job.Name,
			job.ID,
			job.State,
			job.Queue,
			fmt.Sprintf("%s/%s (%d%%)", job.Walltime, job.RequestedWalltime, timePercentage),
			strings.Join(job.Nodes, ", "),
		})
	}
	table.Render()
	return nil
}

func listJobs() ([]Job, error) {
	listOutput, err := query()
	if err != nil {
		return nil, fmt.Errorf("query list: %w", err)
	}

	jobList, err := listOutputToJobList(listOutput)
	if err != nil {
		return nil, fmt.Errorf("convert to job list: %w", err)
	}
	return jobList, nil
}

func list() error {
	jobList, err := listJobs()
	if err != nil {
		return fmt.Errorf("query  job list: %w", err)
	}

	if err := showTable(jobList); err != nil {
		return fmt.Errorf("query list: %w", err)
	}

	return nil
}
