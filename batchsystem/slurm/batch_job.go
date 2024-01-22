package slurm

import (
	"fmt"
	"strings"

	"github.com/unkaktus/robin/batchsystem"
)

func (b *Slurm) JobData(job batchsystem.Job) (string, error) {
	header, err := batchsystem.ExecTemplate(`#!/bin/bash -l
#SBATCH -J {{.Name}}
#SBATCH -o {{.OutputFile}}
#SBATCH -e {{.ErrorFile}}
{{ if ne .Account ""}}#SBATCH --account={{.Account}}{{end}}
{{ if ne .Partition ""}}#SBATCH --partition={{.Partition}}{{end}}
#SBATCH --mail-type=ALL
{{ if ne .Email ""}}#SBATCH --mail-user={{.Email}}{{end}}
#SBATCH --nodes {{.Nodes}}
#SBATCH --ntasks-per-node {{.TasksPerNode}}
#SBATCH --time={{.Walltime}}
{{range .ExtraFlags}}#SBATCH {{.}}
{{end}}
`, job)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	jobData := header

	if job.WorkingDirectory != "" {
		jobData += fmt.Sprintf("cd %s\n", job.WorkingDirectory)
	}

	if len(job.InitScript) > 0 {
		for _, line := range job.InitScript {
			jobData += fmt.Sprintf("%s\n", line)
		}
	}

	task := []string{
		"srun", "robin", "nest",
	}

	task = append(task, job.Runtime...)
	task = append(task, job.Executable)

	for _, argument := range job.Arguments {
		if strings.Contains(argument, "{{.ConfigFilename}}") {
			argument = strings.ReplaceAll(argument, "{{.ConfigFilename}}", job.ConfigFilename)
		}
		task = append(task, argument)
	}

	jobData += strings.Join(task, " ")
	jobData += "\n"

	return jobData, nil
}
