package robin

import (
	"time"

	"github.com/unkaktus/robin/nest"
)

type BatchSystem interface {
	Init() error
	ListJobs(all bool) ([]Job, error)
	Shell(jobName string, nodeID int) error
	Cancel(jobName string) error
	NestVariables() nest.Variables
	Submit(jobData string) error
}

type Job struct {
	Name              string
	ID                string
	Queue             string
	State             string
	ExitCode          int
	CreationTime      time.Time
	StartTime         time.Time
	Nodes             []string
	NodeNumber        int
	CPUNumber         int
	CPUTime           time.Duration
	Walltime          time.Duration
	RequestedWalltime time.Duration
	OutputFile        string
	ErrorFile         string
	WorkingDirectory  string
	Comment           string
}
