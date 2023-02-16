package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

func run() error {
	flag.Parse()

	command := flag.Arg(0)
	switch command {
	case "list":
		state := strings.ToUpper(flag.Arg(1))
		if err := list(state); err != nil {
			return fmt.Errorf("list error: %w", err)
		}
	case "ssh":
		jobName := flag.Arg(1)
		nodeID, err := strconv.Atoi(flag.Arg(2))
		if err != nil {
			return fmt.Errorf("node ID must be an integer")
		}
		if err := ssh(jobName, nodeID); err != nil {
			return fmt.Errorf("list error: %w", err)
		}
	case "logs":
		jobName := flag.Arg(1)
		outputType := flag.Arg(2)
		if err := logs(jobName, outputType); err != nil {
			return fmt.Errorf("logs error: %w", err)
		}
	case "logtail":
		jobName := flag.Arg(1)
		outputType := flag.Arg(2)
		if err := logtail(jobName, outputType); err != nil {
			return fmt.Errorf("logs error: %w", err)
		}
	case "cancel":
		jobName := flag.Arg(1)
		if err := cancel(jobName); err != nil {
			return fmt.Errorf("cancel error: %w", err)
		}
	}

	return nil
}
