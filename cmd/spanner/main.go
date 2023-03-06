package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/unkaktus/spanner"
	"github.com/unkaktus/spanner/batchsystem"
	"github.com/unkaktus/spanner/batchsystem/pbs"
	"github.com/unkaktus/spanner/batchsystem/slurm"
)

func run() (err error) {
	flag.Parse()

	var bs spanner.BatchSystem

	switch batchsystem.DetectBatchSystem() {
	case batchsystem.BatchPBS:
		bs = &pbs.PBS{}
	case batchsystem.BatchSlurm:
		bs = &slurm.Slurm{}
	default:
		return fmt.Errorf("unsupported batch system")
	}

	command := flag.Arg(0)
	switch command {
	case "list":
		state := strings.ToUpper(flag.Arg(1))
		if err := spanner.ListJobs(bs, state); err != nil {
			return fmt.Errorf("list error: %w", err)
		}
	case "tent":
		cmdline := flag.Args()[1:]
		if err := spanner.Tent(bs, cmdline); err != nil {
			return fmt.Errorf("tent: %w", err)
		}
	case "begin":
		if err := spanner.Begin(bs, "begin.toml", flag.Arg(1)); err != nil {
			return fmt.Errorf("begin: %w", err)
		}
	case "ssh":
		jobName := flag.Arg(1)
		nodeIDString := flag.Arg(2)
		nodeID := 0
		if nodeIDString != "" {
			nodeID, err = strconv.Atoi(flag.Arg(2))
			if err != nil {
				return fmt.Errorf("node ID must be an integer")
			}
		}
		if err := bs.SSH(jobName, nodeID); err != nil {
			return fmt.Errorf("list error: %w", err)
		}
	case "logs":
		jobName := flag.Arg(1)
		outputType := flag.Arg(2)
		if err := spanner.Logs(bs, jobName, outputType); err != nil {
			return fmt.Errorf("logs error: %w", err)
		}
	case "logtail":
		jobName := flag.Arg(1)
		outputType := flag.Arg(2)
		if err := spanner.Logtail(bs, jobName, outputType); err != nil {
			return fmt.Errorf("logs error: %w", err)
		}
	case "cancel":
		jobName := flag.Arg(1)
		if err := bs.Cancel(jobName); err != nil {
			return fmt.Errorf("cancel error: %w", err)
		}
	case "clear":
		target := flag.Arg(1)
		if target != "history" {
			break
		}
		if err := bs.ClearHistory(); err != nil {
			return fmt.Errorf("clear hisory error: %w", err)
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
