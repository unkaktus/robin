package paths

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"os"
	"path"
	"strings"
)

func RobinHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	robinHome := path.Join(home, ".robin")
	if err := os.MkdirAll(robinHome, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "cannot create robin home: %v\n", err)
		os.Exit(1)
	}
	return robinHome
}

func JobDataPath() string {
	jobdataPath := path.Join(RobinHome(), "jobdata")
	if err := os.MkdirAll(jobdataPath, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "cannot create job data path: %v\n", err)
		os.Exit(1)
	}
	return jobdataPath
}

func JobDataFilename(jobName, jobData string) string {
	h := sha256.New()
	h.Write([]byte(jobData))
	hash := strings.ToLower(base32.HexEncoding.WithPadding(base32.NoPadding).EncodeToString(h.Sum(nil)[:6]))

	filename := path.Join(JobDataPath(), jobName+"."+hash+".sh")
	return filename
}
