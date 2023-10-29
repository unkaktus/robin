package spanner

import (
	"time"

	"github.com/unkaktus/spanner/batchsystem"
	"github.com/unkaktus/spanner/tent"
)

type BatchSystem interface {
	ListJobs(all bool) ([]Job, error)
	Shell(jobName string, nodeID int) error
	Cancel(jobName string) error
	TentVariables() tent.Variables
	JobData(batchsystem.Job) (string, error)
	Submit(jobData string) error
}

type Job struct {
	Name              string
	ID                string
	Queue             string
	State             string
	ExitCode          int
	CreationTime      time.Time
	Nodes             []string
	NodeNumber        int
	CPUNumber         int
	CPUTime           time.Duration
	Walltime          time.Duration
	RequestedWalltime time.Duration
	OutputFile        string
	ErrorFile         string
	WorkingDirectory  string
}
