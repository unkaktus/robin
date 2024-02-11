package robin

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/unkaktus/tablewriter"
)

func showTable(jobList []Job, full bool) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRoundedStyle()
	if full {
		table.SetHeader([]string{"Name", "ID", "State", "Queue", "Time", "Nodes"})
	} else {
		table.SetHeader([]string{"Name", "State", "Queue", "Time", "Nodes"})
	}

	for _, job := range jobList {
		timeString := ""
		if job.RequestedWalltime == time.Duration(0) {
			timeString = job.Walltime.String()
		} else {
			timePercentage := int(100 * job.Walltime.Seconds() / job.RequestedWalltime.Seconds())
			timeString = fmt.Sprintf("[%d%%] %s/%s",
				timePercentage,
				job.Walltime,
				job.RequestedWalltime,
			)
		}
		if full {
			table.Append([]string{
				job.Name,
				job.ID,
				fmt.Sprintf("%s [%d]", job.State, job.ExitCode),
				job.Queue,
				timeString,
				strings.Join(job.Nodes, ", "),
			})
		} else {
			table.Append([]string{
				job.Name,
				fmt.Sprintf("%s [%d]", job.State, job.ExitCode),
				job.Queue,
				timeString,
				strconv.Itoa(job.NodeNumber),
			})
		}
	}
	table.Render()
	return nil
}

type ListRequest struct {
	All             bool
	Full            bool
	MachineReadable bool
	State           string
}

func ListJobs(bs BatchSystem, request ListRequest) error {
	jobList, err := bs.ListJobs(request.All)
	if err != nil {
		return fmt.Errorf("query job list: %w", err)
	}

	jobMap := map[string]Job{}
	for _, job := range jobList {
		addedJob, ok := jobMap[job.Name]
		if ok {
			if job.CreationTime.Before(addedJob.CreationTime) {
				jobMap[job.Name] = job
			}
		} else {
			jobMap[job.Name] = job
		}
	}

	jobList = []Job{}
	for _, job := range jobMap {
		// Skip the job with other states
		if request.State != "" && job.State != request.State {
			continue
		}
		jobList = append(jobList, job)
	}

	sort.Slice(jobList, func(i, j int) bool {
		return jobList[i].Name < jobList[j].Name
	})

	switch {
	case request.MachineReadable:
		if err := json.NewEncoder(os.Stdout).Encode(jobList); err != nil {
			return fmt.Errorf("encode list: %w", err)
		}
	default:
		if err := showTable(jobList, request.Full); err != nil {
			return fmt.Errorf("show table: %w", err)
		}
	}

	return nil
}

func LatestJob(bs BatchSystem) (*Job, error) {
	jobList, err := bs.ListJobs(false)
	if err != nil {
		return nil, fmt.Errorf("query job list: %w", err)
	}
	sort.Slice(jobList, func(i, j int) bool {
		return jobList[i].CreationTime.After(jobList[j].CreationTime)
	})
	if len(jobList) == 0 {
		return nil, fmt.Errorf("no jobs found")
	}

	return &jobList[0], nil
}

func findJob(b BatchSystem, jobName string) (job Job, err error) {
	jobList, err := b.ListJobs(false)
	if err != nil {
		return job, fmt.Errorf("list jobs: %w", err)
	}
	for _, job = range jobList {
		if job.Name == jobName {
			return job, nil
		}
	}
	return job, fmt.Errorf("job not found")
}
