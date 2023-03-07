package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/unkaktus/spanner"
	"github.com/unkaktus/spanner/batchsystem"
	"github.com/unkaktus/spanner/batchsystem/pbs"
	"github.com/unkaktus/spanner/batchsystem/slurm"
	"github.com/urfave/cli/v2"
)

func run() (err error) {
	var bs spanner.BatchSystem

	switch batchsystem.DetectBatchSystem() {
	case batchsystem.BatchPBS:
		bs = &pbs.PBS{}
	case batchsystem.BatchSlurm:
		bs = &slurm.Slurm{}
	default:
		return fmt.Errorf("unsupported batch system")
	}

	app := &cli.App{
		Name: "spanner",
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "list jobs",
				Action: func(cCtx *cli.Context) error {
					state := strings.ToUpper(flag.Arg(1))
					if err := spanner.ListJobs(bs, state); err != nil {
						return fmt.Errorf("list error: %w", err)
					}
					return nil
				},
			},
			{
				Name:  "logs",
				Usage: "get the logs of a job",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "f",
						Value: false,
						Usage: "tail the logs",
					},
					&cli.IntFlag{
						Name:  "n",
						Value: 30,
						Usage: "number of lines in the tail",
					},
				},
				Action: func(cCtx *cli.Context) error {
					jobName := cCtx.Args().Get(0)
					outputType := cCtx.Args().Get(1)
					switch cCtx.Bool("f") {
					case false:
						if err := spanner.Logs(bs, jobName, outputType); err != nil {
							return fmt.Errorf("logs error: %w", err)
						}
					case true:
						nLines := cCtx.Int("n")
						if err := spanner.Logtail(bs, jobName, outputType, nLines); err != nil {
							return fmt.Errorf("logs error: %w", err)
						}
					}

					return nil
				},
			},
			{
				Name:  "begin",
				Usage: "begin a job",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "f",
						Value: "begin.toml",
						Usage: "path to begin.toml file",
					},
					&cli.BoolFlag{
						Name:  "dry",
						Value: false,
						Usage: "dry run: print the job data only, do not submit",
					},
				},
				Action: func(cCtx *cli.Context) error {
					configFilename := cCtx.Args().Get(0)
					if err := spanner.Begin(bs, cCtx.String("f"), configFilename, cCtx.Bool("dry")); err != nil {
						return fmt.Errorf("begin: %w", err)
					}
					return nil
				},
			},
			{
				Name:  "submit",
				Usage: "submit a job file without magic",

				Action: func(cCtx *cli.Context) error {
					jobDataFilename := cCtx.Args().First()
					jobData, err := ioutil.ReadFile(jobDataFilename)
					if err != nil {
						return fmt.Errorf("read job data file: %w", err)
					}
					if err := bs.Submit(string(jobData)); err != nil {
						return fmt.Errorf("submit: %w", err)
					}
					return nil
				},
			},
			{
				Name:    "cancel",
				Aliases: []string{"stop"},
				Usage:   "cancel jobs",
				Action: func(cCtx *cli.Context) error {
					jobName := cCtx.Args().First()
					if err := bs.Cancel(jobName); err != nil {
						return fmt.Errorf("cancel error: %w", err)
					}
					return nil
				},
			},
			{
				Name:    "ssh",
				Aliases: []string{"shell"},
				Usage:   "login into nodes",
				Action: func(cCtx *cli.Context) error {
					jobName := cCtx.Args().Get(0)
					nodeIDString := cCtx.Args().Get(1)
					nodeID := 0
					if nodeIDString != "" {
						nodeID, err = strconv.Atoi(nodeIDString)
						if err != nil {
							return fmt.Errorf("node ID must be an integer")
						}
					}
					if err := bs.SSH(jobName, nodeID); err != nil {
						return fmt.Errorf("ssh error: %w", err)
					}
					return nil
				},
			},
			{
				Name:  "clear",
				Usage: "clear job history",
				Action: func(cCtx *cli.Context) error {
					target := cCtx.Args().First()
					if target != "history" {
						return fmt.Errorf("unknown target: %s", target)
					}
					if err := bs.ClearHistory(); err != nil {
						return fmt.Errorf("clear hisory error: %w", err)
					}
					return nil
				},
			},
			{
				Name:  "tent",
				Usage: "run tent",
				Action: func(cCtx *cli.Context) error {
					cmdline := append([]string{cCtx.Args().First()}, cCtx.Args().Tail()...)
					if err := spanner.Tent(bs, cmdline); err != nil {
						return fmt.Errorf("tent: %w", err)
					}
					return nil
				},
			},
		},
	}
	return app.Run(os.Args)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
