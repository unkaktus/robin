package pbs

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	// Write job data to a file
	if err := os.MkdirAll(".robin/jobdata", 0700); err != nil {
		return fmt.Errorf("cannot create .robin/jobdata directory: %w", err)
	}

	h := sha256.New()
	h.Write([]byte(jobData))
	hash := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	jobDataFilename, err := filepath.Abs(filepath.Join(".robin/jobdata", hash))
	if err != nil {
		return fmt.Errorf("get absolute path for .robin/jobdata: %w", err)
	}

	if err := os.WriteFile(jobDataFilename, []byte(jobData), 0700); err != nil {
		return fmt.Errorf("write job data file: %w", err)
	}

	comment := robin.Comment{
		JobDataFilename: jobDataFilename,
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
