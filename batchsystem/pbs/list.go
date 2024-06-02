package pbs

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/unkaktus/robin"
)

type ListOutput struct {
	Jobs map[string]struct {
		Name         string `json:"Job_Name"`
		State        string `json:"job_state"`
		Queue        string `json:"queue"`
		CreationTime string `json:"ctime"`
		ExecHosts    string `json:"exec_host"`
		ErrorPath    string `json:"Error_Path"`
		OutputPath   string `json:"Output_Path"`
		VariableList struct {
			WorkDir      string `json:"PBS_O_WORKDIR"`
			RobinComment string `json:"robin_comment"`
		} `json:"Variable_List"`
		ExitStatus    int `json:"Exit_status"`
		ResourcesUsed struct {
			CPUTime  string `json:"cput"`
			Walltime string `json:"walltime"`
		} `json:"resources_used"`
		ResourceList struct {
			NodeNumber int    `json:"nodecounter"`
			CPUNumber  int    `json:"ncpus"`
			Walltime   string `json:"walltime"`
		} `json:"Resource_List"`
	} `json:"Jobs"`
}

func query() (*ListOutput, error) {
	cmd := exec.Command("qstat", "-f", "-F", "json")
	out, err := cmd.Output()
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
		if len(n[0]) == 0 {
			continue
		}
		nodes = append(nodes, n[0])
	}
	return nodes, nil
}

func filePath(path string) string {
	if path == "" {
		return ""
	}
	sp := strings.Split(path, ":")
	if len(sp) != 2 {
		return ""
	}
	return sp[1]
}

func listOutputToJobList(listOutput *ListOutput) (jobs []robin.Job, err error) {
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

		job := robin.Job{
			Name:              listedJob.Name,
			ID:                jobID,
			Queue:             listedJob.Queue,
			State:             listedJob.State,
			ExitCode:          listedJob.ExitStatus,
			CreationTime:      creationTime,
			Nodes:             nodes,
			NodeNumber:        listedJob.ResourceList.NodeNumber,
			CPUNumber:         listedJob.ResourceList.CPUNumber,
			CPUTime:           cpuTime,
			Walltime:          walltime,
			RequestedWalltime: requestedWalltime,
			OutputFile:        filePath(listedJob.OutputPath),
			ErrorFile:         filePath(listedJob.ErrorPath),
			WorkingDirectory:  listedJob.VariableList.WorkDir,
			Comment:           listedJob.VariableList.RobinComment,
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (b *PBS) ListJobs(all bool) ([]robin.Job, error) {
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

func (b *PBS) FindJob(jobName string) (*robin.Job, error) {
	jobList, err := b.ListJobs(false)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	for _, job := range jobList {
		if job.Name == jobName {
			return &job, nil
		}
	}
	return nil, nil
}
