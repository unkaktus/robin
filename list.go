package spanner

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/unkaktus/tablewriter"
)

func showTable(jobList []Job) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRoundedStyle()
	table.SetHeader([]string{"Name", "State", "Queue", "Time", "Nodes", "MPI"})

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

		table.Append([]string{
			job.Name,
			fmt.Sprintf("%s [%d]", job.State, job.ExitCode),
			job.Queue,
			timeString,
			strconv.Itoa(job.NodeNumber),
			fmt.Sprintf("%d/%d", job.MPIProcessNumber/job.NodeNumber, job.MPIProcessNumber),
		})
	}
	table.Render()
	return nil
}

func ListJobs(bs BatchSystem, state string) error {
	jobList, err := bs.ListJobs()
	if err != nil {
		return fmt.Errorf("query job list: %w", err)
	}

	jobMap := map[string]Job{}
	for _, job := range jobList {
		addedJob, ok := jobMap[job.Name]
		if ok {
			if job.CreationTime.After(addedJob.CreationTime) {
				jobMap[job.Name] = job
			}
		} else {
			jobMap[job.Name] = job
		}
	}

	jobList = []Job{}
	for _, job := range jobMap {
		// Skip the job with other states
		if state != "" && job.State != state {
			continue
		}
		jobList = append(jobList, job)
	}

	sort.Slice(jobList, func(i, j int) bool {
		return jobList[i].Name < jobList[j].Name
	})

	if err := showTable(jobList); err != nil {
		return fmt.Errorf("query list: %w", err)
	}

	return nil
}

func LatestJob(bs BatchSystem) (*Job, error) {
	jobList, err := bs.ListJobs()
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
