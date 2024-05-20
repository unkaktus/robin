package robin

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hpcloud/tail"
	"github.com/rs/zerolog"
)

func Logs(b BatchSystem, jobName string, outputType string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	job, err := findJob(b, jobName)
	if err != nil {
		return fmt.Errorf("find job: %w", err)
	}

	logFile := job.OutputFile
	if outputType == "err" {
		logFile = job.ErrorFile
	}
	cmd := exec.Command(editor, logFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute tail: %w", err)
	}

	return nil
}

func writeLine(consoleWriter zerolog.ConsoleWriter, line string) error {
	_, err := consoleWriter.Write([]byte(line))
	if err != nil {
		_, err = consoleWriter.Out.Write([]byte(line + "\n"))
		if err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	return nil
}

func Logtail(b BatchSystem, jobName, outputType string, nBytes int) error {
	job, err := findJob(b, jobName)
	if err != nil {
		return fmt.Errorf("find job: %w", err)
	}

	logFile := job.OutputFile
	if outputType == "err" {
		logFile = job.ErrorFile
	}

	tailConfig := tail.Config{
		Follow: true,
		ReOpen: true,
		Poll:   true, // On many cluster filesystems, inotify doesn't work
		Location: &tail.SeekInfo{
			Offset: -int64(nBytes),
			Whence: io.SeekEnd,
		},
		Logger: tail.DiscardingLogger,
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out: os.Stdout,
		PartsOrder: []string{
			zerolog.TimestampFieldName,
			zerolog.LevelFieldName,
			"stream",
			zerolog.MessageFieldName,
		},
		TimeFormat:    time.DateTime,
		FieldsExclude: []string{"stream"},
	}
	t, err := tail.TailFile(logFile, tailConfig)
	if err != nil {
		return fmt.Errorf("tail file: %w", err)
	}

	for line := range t.Lines {
		if line.Err == io.EOF {
			return nil
		}
		if line.Err != nil {
			return fmt.Errorf("tail file: %w", err)
		}
		// If it is a cut JSON, skip it
		if strings.HasSuffix(line.Text, "}") && !strings.HasPrefix(line.Text, "{") {
			continue
		}
		if err = writeLine(consoleWriter, line.Text); err != nil {
			return fmt.Errorf("write line: %w", err)
		}
	}

	return nil
}
