package batchsystem

import (
	"fmt"
	"strings"
	"text/template"
)

type Job struct {
	Name           string
	ConfigFilename string

	Nodes        int
	TasksPerNode int

	NodeType  string
	Partition string
	Account   string
	Walltime  string
	Email     string

	ExtraFlags []string

	WorkingDirectory string
	OutputFile       string
	ErrorFile        string

	InitScript []string
	Runtime    []string
	Executable string
	Arguments  []string
}

func ExecTemplate(ts string, s interface{}) (string, error) {
	t, err := template.New("template").Parse(ts)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)

	}
	builder := &strings.Builder{}

	err = t.Execute(builder, s)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return builder.String(), nil
}
