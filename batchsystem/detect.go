package batchsystem

import "os/exec"

func DetectBatchSystem() string {
	if _, err := exec.LookPath("qstat"); err == nil {
		return BatchPBS
	}
	if _, err := exec.LookPath("squeue"); err == nil {
		return BatchSlurm
	}
	if _, err := exec.LookPath("tmux"); err == nil {
		return BatchTmux
	}
	return BatchUnsupported
}
