package tmux

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/unkaktus/robin"
)

type ListOutput []struct {
	Name             string `json:"name"`
	ID               string `json:"id"`
	WorkingDirectory string `json:"working_directory"`
	CreationTime     int64  `json:"creation_time"`
}

func query() (ListOutput, error) {
	format := `{"name": "#{session_name}", "id": "#{session_id}", "working_directory": "#{session_path}", "creation_time": #{session_created} },`
	cmd := exec.Command("tmux", "list-sessions", "-F", format)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("exectute command: %w", err)
	}
	outList := "[" + strings.TrimRight(string(out), ",\n") + "]"

	listOutput := ListOutput{}

	err = json.Unmarshal([]byte(outList), &listOutput)
	if err != nil {
		return nil, fmt.Errorf("unmarshal JSON: %w", err)
	}
	return listOutput, nil
}

func listOutputToJobList(listOutput ListOutput) (jobs []robin.Job, err error) {
	for _, listedJob := range listOutput {
		if !strings.HasPrefix(listedJob.Name, "robin_") {
			continue
		}
		nameData := &NameData{}
		err = nameData.DecodeString(listedJob.Name)
		if err != nil {
			return nil, fmt.Errorf("decode name data: %w", err)
		}

		creationTime := time.Unix(listedJob.CreationTime, 0)

		job := robin.Job{
			Name:             nameData.Name,
			ID:               nameData.EncodeToString(),
			Queue:            "tmux",
			State:            "R",
			CreationTime:     creationTime,
			StartTime:        creationTime,
			NodeNumber:       1,
			Nodes:            []string{"localhost"},
			Walltime:         time.Since(creationTime),
			OutputFile:       nameData.LogFile,
			WorkingDirectory: listedJob.WorkingDirectory,
			Comment:          nameData.Comment.Encode(),
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (b *Tmux) ListJobs(all bool) ([]robin.Job, error) {
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

func (b *Tmux) FindJob(jobName string) (*robin.Job, error) {
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
