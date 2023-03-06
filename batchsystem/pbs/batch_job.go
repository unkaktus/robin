package pbs

import (
	"fmt"

	"github.com/unkaktus/spanner/batchsystem"
)

func (b *PBS) JobData(job batchsystem.Job) (string, error) {
	_, err := batchsystem.ExecTemplate(`#!/bin/bash -l
#PBS -N {{.Name}}
#PBS -e {{.ErrorFile}}
#PBS -o {{.OutputFile}}
#PBS -m abe
#PBS -M {{.Email}}
#PBS -l select={{.NumberOfNodes}}`+
		`:node_type={{.NodeType}}`+
		`:mpiprocs={{.NumberOfMPIRanksPerNode}}`+
		`:ompthreads={{.NumberOfOMPThreadsPerProcess}}`+`
#PBS -l walltime={{.WalltimeString}}`,
		job)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return "", fmt.Errorf("not implemented")
}
