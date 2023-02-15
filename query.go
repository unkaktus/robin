package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ListOutput struct {
	Jobs map[string]struct {
		Name          string `json:"Job_Name"`
		State         string `json:"job_state"`
		Queue         string `json:"queue"`
		ExecHosts     string `json:"exec_host"`
		ResourcesUsed struct {
			CPUTime  string `json:"cput"`
			Walltime string `json:"walltime"`
		} `json:"resources_used"`
		ResourceList struct {
			Walltime string `json:"walltime"`
		} `json:"Resource_List"`
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

const clockLayout = "15:04:05"

var clockZero, _ = time.Parse(clockLayout, "00:00:00")

func clockDuration(clock string) (time.Duration, error) {
	c, err := time.Parse(clockLayout, clock)
	if err != nil {
		return 0, err
	}
	return c.Sub(clockZero), nil
}

func parseNodeList(s string) ([]string, error) {
	nodes := []string{}
	sp := strings.Split(s, "+")
	for _, nd := range sp {
		n := strings.Split(nd, "/")
		nodes = append(nodes, n[0])
	}
	return nodes, nil
}

type Job struct {
	Name              string
	ID                string
	Queue             string
	State             string
	Nodes             []string
	CPUTime           time.Duration
	Walltime          time.Duration
	RequestedWalltime time.Duration
}

func listOutputToJobList(listOutput *ListOutput) (jobs []Job, err error) {
	for jobID, listedJob := range listOutput.Jobs {
		nodes, err := parseNodeList(listedJob.ExecHosts)
		if err != nil {
			return nil, fmt.Errorf("parsing node list: %w", err)
		}

		var cpuTime time.Duration
		if listedJob.ResourcesUsed.CPUTime != "" {
			cpuTime, err = clockDuration(listedJob.ResourcesUsed.CPUTime)
			if err != nil {
				return nil, fmt.Errorf("parsing CPUTime: %w", err)
			}
		}
		var walltime time.Duration
		if listedJob.ResourcesUsed.Walltime != "" {
			walltime, err = clockDuration(listedJob.ResourcesUsed.Walltime)
			if err != nil {
				return nil, fmt.Errorf("parsing Walltime: %w", err)
			}
		}

		requestedWalltime, err := clockDuration(listedJob.ResourceList.Walltime)
		if err != nil {
			return nil, fmt.Errorf("parsing RequestedWalltime: %w", err)
		}

		job := Job{
			Name:              listedJob.Name,
			ID:                jobID,
			Queue:             listedJob.Queue,
			State:             listedJob.State,
			Nodes:             nodes,
			CPUTime:           cpuTime,
			Walltime:          walltime,
			RequestedWalltime: requestedWalltime,
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}
