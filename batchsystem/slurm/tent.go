package slurm

import (
	"fmt"
	"os"
	"strconv"

	"github.com/unkaktus/spanner/tent"
)

func getTaskID() (int, error) {
	nodeID, err := strconv.Atoi(os.Getenv("SLURM_NODEID"))
	if err != nil {
		return 0, fmt.Errorf("parsing SLURM_NODEID: %w", err)
	}
	tasksPerNode, err := strconv.Atoi(os.Getenv("SLURM_NTASKS_PER_NODE"))
	if err != nil {
		return 0, fmt.Errorf("parsing SLURM_NTASKS_PER_NODE: %w", err)
	}
	localID, err := strconv.Atoi(os.Getenv("SLURM_LOCALID"))
	if err != nil {
		return 0, fmt.Errorf("parsing SLURM_LOCALID: %w", err)
	}
	taskID := nodeID*tasksPerNode + localID
	return taskID, nil
}

func getTotalTaskNumber() (totalTaskNumber int, err error) {
	totalTaskNumber, err = strconv.Atoi(os.Getenv("SLURM_NTASKS"))
	if err != nil {
		return 0, fmt.Errorf("parsing SLURM_NTASKS: %w", err)
	}

	return totalTaskNumber, nil
}

func (b *Slurm) TentVariables() tent.Variables {
	vars := tent.Variables{}
	if taskID, err := getTaskID(); err == nil {
		vars.TaskID = taskID
	}
	if totalTaskNumber, err := getTotalTaskNumber(); err == nil {
		vars.TotalTaskNumber = totalTaskNumber
	}

	return vars
}
