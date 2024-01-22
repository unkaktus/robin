package pbs

import (
	"testing"

	"github.com/matryer/is"
)

func TestJobNameFromJobData(t *testing.T) {
	is := is.New(t)

	jobData := `#!/bin/bash -l
#PBS -o robin.out
#PBS -l select=3:node_type=rome:ncpus=128:mpiprocs=16:ompthreads=8
#PBS -l walltime=12:00:00
#PBS -N robin_job_name


mpirun -n 72 robin nest application`

	jobName := jobNameFromJobData(jobData)
	is.Equal(jobName, "robin_job_name")
}
