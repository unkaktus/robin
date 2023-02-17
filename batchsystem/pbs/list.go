package pbs

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/unkaktus/spanner"
)

type ListOutput struct {
	Jobs map[string]struct {
		Name          string `json:"Job_Name"`
		State         string `json:"job_state"`
		Queue         string `json:"queue"`
		CreationTime  string `json:"ctime"`
		ExecHosts     string `json:"exec_host"`
		ErrorPath     string `json:"Error_Path"`
		OutputPath    string `json:"Output_Path"`
		ResourcesUsed struct {
			CPUTime  string `json:"cput"`
			Walltime string `json:"walltime"`
		} `json:"resources_used"`
		ResourceList struct {
			NodeNumber       int    `json:"nodecounter"`
			CPUNumber        int    `json:"ncpus"`
			MPIProcessNumber int    `json:"mpiprocs"`
			Walltime         string `json:"walltime"`
		} `json:"Resource_List"`
	} `json:"Jobs"`
}

func query() (*ListOutput, error) {
	cmd := exec.Command("qstat", "-xf", "-F", "json")
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

func clockDuration(clock string) (time.Duration, error) {
	sp := strings.Split(clock, ":")
	if len(sp) != 3 {
		return 0, fmt.Errorf("wrong string length")
	}
	h, m, s := sp[0], sp[1], sp[2]
	d, err := time.ParseDuration(fmt.Sprintf("%sh%sm%ss", h, m, s))
	if err != nil {
		return 0, fmt.Errorf("parse duration: %w", err)
	}
	return d, nil
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

func listOutputToJobList(listOutput *ListOutput) (jobs []spanner.Job, err error) {
	for jobID, listedJob := range listOutput.Jobs {
		var creationTime time.Time
		if listedJob.CreationTime != "" {
			creationTime, err = time.Parse(time.ANSIC, listedJob.CreationTime)
			if err != nil {
				return nil, fmt.Errorf("parsing creation time: %w", err)
			}
		}

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

		job := spanner.Job{
			Name:              listedJob.Name,
			ID:                jobID,
			Queue:             listedJob.Queue,
			State:             listedJob.State,
			CreationTime:      creationTime,
			Nodes:             nodes,
			NodeNumber:        listedJob.ResourceList.NodeNumber,
			CPUNumber:         listedJob.ResourceList.CPUNumber,
			MPIProcessNumber:  listedJob.ResourceList.MPIProcessNumber,
			CPUTime:           cpuTime,
			Walltime:          walltime,
			RequestedWalltime: requestedWalltime,
			OutputFile:        strings.Split(listedJob.OutputPath, ":")[1],
			ErrorFile:         strings.Split(listedJob.ErrorPath, ":")[1],
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (b *PBS) ListJobs() ([]spanner.Job, error) {
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
