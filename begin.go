package spanner

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/unkaktus/spanner/batchsystem"
	"github.com/unkaktus/spanner/begin"
)

func Begin(b BatchSystem, beginFilename, configFilename string) error {
	beginfile, err := begin.ParseBeginfile(beginFilename)
	if err != nil {
		return fmt.Errorf("parse beginfile: %w", err)
	}
	configName := strings.TrimSuffix(
		filepath.Base(configFilename),
		filepath.Ext(configFilename),
	)

	name := beginfile.Name + "_" + configName

	outputFile := filepath.Join(
		beginfile.WorkingDirectory,
		fmt.Sprintf("%s.out", name),
	)
	errorFile := filepath.Join(
		beginfile.WorkingDirectory,
		fmt.Sprintf("%s.err", name),
	)

	job := batchsystem.Job{
		Name:             name,
		ConfigFilename:   configFilename,
		OutputFile:       outputFile,
		ErrorFile:        errorFile,
		Nodes:            beginfile.Nodes,
		TasksPerNode:     beginfile.TasksPerNode,
		NodeType:         beginfile.NodeType,
		Walltime:         begin.FormatDuration(beginfile.Walltime),
		Email:            beginfile.Email,
		WorkingDirectory: beginfile.WorkingDirectory,
		InitScript:       beginfile.InitScript,
		Runtime:          beginfile.Runtime,
		Executable:       beginfile.Executable,
		Arguments:        beginfile.Arguments,
	}

	jobData, err := b.JobData(job)
	if err != nil {
		return fmt.Errorf("get job data: %w", err)
	}
	fmt.Printf("%s", jobData)
	// Submit
	return nil
}
