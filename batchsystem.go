package spanner

import (
	"time"

	"github.com/unkaktus/spanner/batchsystem"
	"github.com/unkaktus/spanner/tent"
)

type BatchSystem interface {
	ListJobs() ([]Job, error)
	SSH(jobName string, nodeID int) error
	ClearHistory() error
	Cancel(jobName string) error
	TentVariables() tent.Variables
	JobData(batchsystem.Job) (string, error)
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
	MPIProcessNumber  int
	CPUTime           time.Duration
	Walltime          time.Duration
	RequestedWalltime time.Duration
	OutputFile        string
	ErrorFile         string
}
