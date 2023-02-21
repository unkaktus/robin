package spanner

import (
	"os/exec"
	"time"
)

const (
	BatchPBS         = "pbs"
	BatchSlurm       = "slurm"
	BatchUnsupported = "unsupported"
)

func DetectBatchSystem() string {
	if _, err := exec.LookPath("qstat"); err == nil {
		return BatchPBS
	}
	if _, err := exec.LookPath("squeue"); err == nil {
		return BatchSlurm
	}
	return BatchUnsupported
}

type BatchSystem interface {
	ListJobs() ([]Job, error)
	SSH(jobName string, nodeID int) error
	ClearHistory() error
	Logs(jobName string, outputType string) error
	Logtail(jobName string, outputType string) error
	Cancel(jobName string) error
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
