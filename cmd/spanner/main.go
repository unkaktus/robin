package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/unkaktus/spanner"
	"github.com/unkaktus/spanner/batchsystem"
	"github.com/unkaktus/spanner/batchsystem/pbs"
	"github.com/unkaktus/spanner/batchsystem/slurm"
	"github.com/urfave/cli/v2"
)

var (
	errUnsupported error = errors.New("unsupported batch system")
)

func run() (err error) {
	var bs spanner.BatchSystem

	switch batchsystem.DetectBatchSystem() {
	case batchsystem.BatchPBS:
		bs = &pbs.PBS{}
	case batchsystem.BatchSlurm:
		bs = &slurm.Slurm{}
	}

	app := &cli.App{
		Name:     "spanner",
		HelpName: "spanner",
		Usage:    "One tool for all HPC",
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Ivan Markin",
				Email: "git@unkaktus.art",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "list jobs",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "all",
						Value: false,
						Usage: "include jobs of other users",
					},
					&cli.BoolFlag{
						Name:  "full",
						Value: false,
						Usage: "include more information on the job, e.g. job ID and node list",
					},
					&cli.StringFlag{
						Name:  "state",
						Value: "",
						Usage: "select the jobs with certain state",
					},
					&cli.BoolFlag{
						Name:  "json",
						Value: false,
						Usage: "output the list in JSON format",
					}},
				Action: func(cCtx *cli.Context) error {
					if bs == nil {
						return errUnsupported
					}

					listRequest := spanner.ListRequest{
						All:             cCtx.Bool("all"),
						Full:            cCtx.Bool("full"),
						MachineReadable: cCtx.Bool("json"),
						State:           strings.ToUpper(cCtx.String("state")),
					}
					if err := spanner.ListJobs(bs, listRequest); err != nil {
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
					&cli.BoolFlag{
						Name:  "latest",
						Value: false,
						Usage: "use the latest running job",
					},
					&cli.BoolFlag{
						Name:  "err",
						Value: false,
						Usage: "output error logs",
					},
				},
				Action: func(cCtx *cli.Context) error {
					if bs == nil {
						return errUnsupported
					}
					jobName := cCtx.Args().Get(0)
					if cCtx.Bool("latest") {
						job, err := spanner.LatestJob(bs)
						if err != nil {
							return fmt.Errorf("looking up the latest job: %w", err)
						}
						jobName = job.Name
					}
					outputType := "out"
					if cCtx.Bool("err") {
						outputType = "err"
					}
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
					if bs == nil {
						return errUnsupported
					}

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
					if bs == nil {
						return errUnsupported
					}

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
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "latest",
						Value: false,
						Usage: "use the latest running job",
					},
				},
				Action: func(cCtx *cli.Context) error {
					if bs == nil {
						return errUnsupported
					}

					jobName := cCtx.Args().First()
					if cCtx.Bool("latest") {
						job, err := spanner.LatestJob(bs)
						if err != nil {
							return fmt.Errorf("looking up the latest job: %w", err)
						}
						jobName = job.Name
					}
					if err := bs.Cancel(jobName); err != nil {
						return fmt.Errorf("cancel error: %w", err)
					}
					return nil
				},
			},
			{
				Name:  "shell",
				Usage: "login into nodes",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "latest",
						Value: false,
						Usage: "use the latest running job",
					},
				},
				Action: func(cCtx *cli.Context) error {
					if bs == nil {
						return errUnsupported
					}

					jobName := cCtx.Args().Get(0)
					if cCtx.Bool("latest") {
						job, err := spanner.LatestJob(bs)
						if err != nil {
							return fmt.Errorf("looking up the latest job: %w", err)
						}
						jobName = job.Name
					}
					nodeIDString := cCtx.Args().Get(1)
					nodeID := 0
					if nodeIDString != "" {
						nodeID, err = strconv.Atoi(nodeIDString)
						if err != nil {
							return fmt.Errorf("node ID must be an integer")
						}
					}
					if err := bs.Shell(jobName, nodeID); err != nil {
						return fmt.Errorf("ssh error: %w", err)
					}
					return nil
				},
			},
			{
				Name:  "tent",
				Usage: "run tent",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "split-output",
						Value: false,
						Usage: "use separate stderr and stdout",
					},
					&cli.BoolFlag{
						Name:  "no-command",
						Value: false,
						Usage: "start tent without any command",
					},
				},
				Action: func(cCtx *cli.Context) error {
					if bs == nil {
						return errUnsupported
					}

					cmdline := append([]string{cCtx.Args().First()}, cCtx.Args().Tail()...)
					mergeOutput := !cCtx.Bool("split-output")
					noCommand := cCtx.Bool("no-command")
					if err := spanner.Tent(bs, cmdline, mergeOutput, noCommand); err != nil {
						return fmt.Errorf("tent: %w", err)
					}
					return nil
				},
			},
			{
				Name:  "on",
				Usage: "run spanner commands remotely, e.g., spanner on machine list",
				Action: func(cCtx *cli.Context) error {
					machine := cCtx.Args().First()
					cmdline := cCtx.Args().Tail()
					if err := spanner.On(machine, cmdline); err != nil {
						return fmt.Errorf("spanner on remote: %w", err)
					}
					return nil
				},
			},
			{
				Name:  "port-forward",
				Usage: "forward a TCP port to a job node on a cluster",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Value:   0,
						Usage:   "port to forward",
					},
					&cli.StringFlag{
						Name:    "machine",
						Aliases: []string{"m"},
						Value:   "",
						Usage:   "machine to connect to, i.e. login node",
					},
				},
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
					port := cCtx.Int("port")
					if port == 0 {
						return fmt.Errorf("port must be specified")
					}
					machine := cCtx.String("machine")
					if machine == "" {
						return fmt.Errorf("machine must be specified")
					}
					if err := spanner.PortForward(machine, jobName, port, nodeID); err != nil {
						return fmt.Errorf("ssh error: %w", err)
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
		fmt.Fprintf(os.Stderr, "spanner: %v\n", err)
		os.Exit(1)
	}
}
