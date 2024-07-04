package robin

import (
	"fmt"
	"os"
	"os/exec"
)

func Show(b BatchSystem, jobName string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	job, err := findJob(b, jobName)
	if err != nil {
		return fmt.Errorf("find job: %w", err)
	}

	comment := &Comment{}
	if err := comment.Decode(job.Comment); err != nil {
		return fmt.Errorf("decode comment: %w", err)
	}

	if comment.JobDataFilename == "" {
		return fmt.Errorf("no job data file")
	}

	cmd := exec.Command(editor, comment.JobDataFilename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute editor: %w", err)
	}

	return nil
}
