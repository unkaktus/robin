package slurm

import (
	"testing"

	"github.com/matryer/is"
)

func TestJobNameFromJobData(t *testing.T) {
	is := is.New(t)

	jobData := `#!/bin/bash
	#SBATCH -o robin.out
	#SBATCH -n 72
	#SBATCH -t 24:00:00
	#SBATCH --mem=256000

	#SBATCH -J robin_job_name

	robin nest mpirun -n 72 application`

	jobName := jobNameFromJobData(jobData)
	is.Equal(jobName, "robin_job_name")
}
