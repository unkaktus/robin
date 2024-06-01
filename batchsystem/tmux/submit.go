package tmux

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

func parseJobData(jobData string) map[string]string {
	m := map[string]string{}
	lines := strings.Split(jobData, "\n")
	for _, line := range lines {
		words := strings.FieldsFunc(line, func(c rune) bool {
			return unicode.IsSpace(c) || c == '='
		})
		if len(words) == 0 {
			continue
		}
		if words[0] != "#TMUX" {
			continue
		}
		if len(words) != 3 {
			continue
		}
		key := strings.TrimPrefix(words[1], "--")
		m[key] = words[2]
	}
	return m
}

func (b *Tmux) Submit(jobData string) error {
	jobParameters := parseJobData(jobData)

	if _, ok := jobParameters["job-name"]; !ok {
		return fmt.Errorf("job has no name")
	}

	if _, ok := jobParameters["log-file"]; !ok {
		return fmt.Errorf("job has no log file set")
	}

	logFile, err := filepath.Abs(jobParameters["log-file"])
	if err != nil {
		return fmt.Errorf("get absolute path to the log file: %w", err)
	}

	job, err := b.FindJob(jobParameters["job-name"])
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

	wrapperScript := fmt.Sprintf("robin nest %s >> %s", jobDataFilename, jobParameters["log-file"])

	nameData := NameData{
		Name:    jobParameters["job-name"],
		LogFile: logFile,
	}

	jobDataReader := strings.NewReader(jobData)
	cmd := exec.Command("tmux", "new-session", "-d", "-s", nameData.EncodeToString(), wrapperScript)
	cmd.Stdin = jobDataReader
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute qsub: %w", err)
	}

	return nil
}
