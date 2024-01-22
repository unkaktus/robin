package pbs

import (
	"fmt"

	"github.com/unkaktus/robin/batchsystem"
)

func (b *PBS) JobData(job batchsystem.Job) (string, error) {
	_, err := batchsystem.ExecTemplate(`#!/bin/bash -l
#PBS -N {{.Name}}
#PBS -e {{.ErrorFile}}
#PBS -o {{.OutputFile}}
#PBS -m abe
{{ if ne .Email ""}}#PBS -M {{.Email}}{{end}}
#PBS -l select={{.NumberOfNodes}}`+
		`:node_type={{.NodeType}}`+`
#PBS -l walltime={{.WalltimeString}}
{{range .ExtraFlags}}#PBS {{.}}
{{end}}
`, job)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return "", fmt.Errorf("not implemented")
}
