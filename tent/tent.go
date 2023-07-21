package tent

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"strings"
)

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

type Variables struct {
	ConfigFilename  string
	TaskID          int
	TotalTaskNumber int
}

func RunCommand(cmdline []string, vars Variables) (process *os.Process, err error) {
	for i, arg := range cmdline {
		cmdline[i], err = ExecTemplate(arg, vars)
		if err != nil {
			return nil, fmt.Errorf("executing template on argument %s: %w", arg, err)
		}
	}
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("task start: %v", err)
	}
	return cmd.Process, nil
}
