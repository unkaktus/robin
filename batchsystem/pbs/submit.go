package pbs

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/unkaktus/robin"
)

func jobNameFromJobData(jobData string) string {
	lines := strings.Split(jobData, "\n")
	for _, line := range lines {
		words := strings.FieldsFunc(line, func(c rune) bool {
			return unicode.IsSpace(c) || c == '='
		})
		if len(words) == 0 {
			continue
		}
		if words[0] != "#PBS" {
			continue
		}
		if len(words) != 3 {
			continue
		}
		if words[1] != "-N" {
			continue
		}
		return words[2]
	}
	return ""
}

func (b *PBS) Submit(jobData string) error {
	jobName := jobNameFromJobData(jobData)
	if jobName == "" {
		return fmt.Errorf("job has no name")
	}

	job, err := b.FindJob(jobName)
	if err != nil {
		return fmt.Errorf("find job: %w", err)
	}

	if job != nil {
		return fmt.Errorf("job with this name already exists")
	}

	comment := robin.Comment{
		JobData: base64.RawStdEncoding.EncodeToString([]byte(jobData)),
	}

	jobDataReader := strings.NewReader(jobData)
	cmd := exec.Command("qsub", "-v", fmt.Sprintf("robin_comment=%s", comment.Encode()))
	cmd.Stdin = jobDataReader
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute qsub: %w", err)
	}
	return nil
}
