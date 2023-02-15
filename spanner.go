package main

import (
	"flag"
	"fmt"
)

func run() error {
	flag.Parse()

	command := flag.Arg(0)
	switch command {
	case "list":
		if err := list(); err != nil {
			return fmt.Errorf("list error: %w", err)
		}
	}

	return nil
}
