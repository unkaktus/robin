package tent

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
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

type Process struct {
	cmd            *exec.Cmd
	wg             sync.WaitGroup
	stdout, stderr io.ReadCloser
}

func (p *Process) Wait() error {
	err := p.cmd.Wait()
	p.stdout.Close()
	p.stderr.Close()
	p.wg.Wait()
	return err
}

func RunCommand(cmdline []string, vars Variables) (*Process, error) {
	for i, arg := range cmdline {
		var err error
		cmdline[i], err = ExecTemplate(arg, vars)
		if err != nil {
			return nil, fmt.Errorf("executing template on argument %s: %w", arg, err)
		}
	}
	cmd := exec.Command(cmdline[0], cmdline[1:]...)

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	process := &Process{
		cmd:    cmd,
		stdout: stdoutReader,
		stderr: stderrReader,
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	process.wg.Add(1)
	go func() {
		defer process.wg.Done()
		logger := log.Logger.With().Str("stream", "stdout").Logger()
		scanner := bufio.NewScanner(process.stdout)
		for scanner.Scan() {
			logger.Info().Msg(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			logger.Err(err)
		}

	}()

	process.wg.Add(1)
	go func() {
		defer process.wg.Done()
		logger := log.Logger.With().Str("stream", "stderr").Logger()
		scanner := bufio.NewScanner(process.stderr)
		for scanner.Scan() {
			logger.Info().Msg(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			logger.Err(err)
		}
	}()

	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("task start: %v", err)
	}
	return process, nil
}
