package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/minio/selfupdate"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/unkaktus/robin"
	"github.com/unkaktus/robin/batchsystem"
	"github.com/unkaktus/robin/batchsystem/pbs"
	"github.com/unkaktus/robin/batchsystem/slurm"
	"github.com/unkaktus/robin/batchsystem/tmux"
	"github.com/urfave/cli/v2"
)

var (
	version        string
	errUnsupported error = errors.New("unsupported batch system")
)

func run() (err error) {
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	var bs robin.BatchSystem

	switch batchsystem.DetectBatchSystem() {
	case batchsystem.BatchPBS:
		bs = &pbs.PBS{}
	case batchsystem.BatchSlurm:
		bs = &slurm.Slurm{}
	case batchsystem.BatchTmux:
		bs = &tmux.Tmux{}
	}
	if err := bs.Init(); err != nil {
		log.Fatal().Err(err).Msg("initialize tmux")
	}

	app := &cli.App{
		Name:     "robin",
		HelpName: "robin",
		Usage:    "One tool for all HPC",
		Authors: []*cli.Author{
			{
				Name:  "Ivan Markin",
				Email: "git@unkaktus.art",
			},
		},
		Version: version,
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
					},
				},
				Action: func(cCtx *cli.Context) error {
					if bs == nil {
						return errUnsupported
					}

					listRequest := robin.ListRequest{
						All:             cCtx.Bool("all"),
						Full:            cCtx.Bool("full"),
						MachineReadable: cCtx.Bool("json"),
						State:           strings.ToUpper(cCtx.String("state")),
					}
					if err := robin.ListJobs(bs, listRequest); err != nil {
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
						Value: 5120,
						Usage: "number of bytes from the end to show, 0 to start from the beginning",
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
						job, err := robin.LatestJob(bs)
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
						if err := robin.Logs(bs, jobName, outputType); err != nil {
							return fmt.Errorf("logs error: %w", err)
						}
					case true:
						nBytes := cCtx.Int("n")
						if err := robin.Logtail(bs, jobName, outputType, nBytes); err != nil {
							return fmt.Errorf("logs error: %w", err)
						}
					}

					return nil
				},
			},
			{
				Name:  "submit",
				Usage: "submit a job",
				Action: func(cCtx *cli.Context) error {
					if bs == nil {
						return errUnsupported
					}

					name := cCtx.Args().First()
					if err := robin.Submit(bs, name); err != nil {
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
						job, err := robin.LatestJob(bs)
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
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Value:   false,
						Usage:   "print errors",
					},
				},
				Action: func(cCtx *cli.Context) error {
					if bs == nil {
						return errUnsupported
					}

					jobName := cCtx.Args().Get(0)
					if cCtx.Bool("latest") {
						job, err := robin.LatestJob(bs)
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
					verbose := cCtx.Bool("verbose")
					return bs.Shell(jobName, nodeID, verbose)
				},
			},
			{
				Name:  "nest",
				Usage: "run nest",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "no-command",
						Value: false,
						Usage: "start nest without any command",
					},
					&cli.BoolFlag{
						Name:  "verbose",
						Value: false,
						Usage: "print misc errors (shell, node head)",
					},
				},
				Action: func(cCtx *cli.Context) error {
					if bs == nil {
						return errUnsupported
					}

					cmdline := append([]string{cCtx.Args().First()}, cCtx.Args().Tail()...)
					noCommand := cCtx.Bool("no-command")
					verbose := cCtx.Bool("verbose")
					if err := robin.Nest(bs, cmdline, noCommand, verbose); err != nil {
						return fmt.Errorf("nest: %w", err)
					}
					return nil
				},
			},
			{
				Name:  "proxy",
				Usage: "run service proxy",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "addr",
						Value: "localhost:9100",
						Usage: "",
					},
				},
				Action: func(cCtx *cli.Context) error {
					if bs == nil {
						return errUnsupported
					}
					addr := cCtx.String("addr")

					if err := robin.Proxy(bs, addr); err != nil {
						return fmt.Errorf("nest: %w", err)
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
					if err := robin.PortForward(machine, jobName, port, nodeID); err != nil {
						return fmt.Errorf("ssh error: %w", err)
					}
					return nil
				},
			},
			{
				Name:  "update",
				Usage: "update itself",
				Action: func(cCtx *cli.Context) error {
					robinURL := fmt.Sprintf("https://github.com/unkaktus/robin/releases/latest/download/robin-%s-%s", runtime.GOOS, runtime.GOARCH)
					resp, err := http.Get(robinURL)
					if err != nil {
						return fmt.Errorf("download release binary: %w", err)
					}
					if resp.StatusCode != http.StatusOK {
						return fmt.Errorf("unsuccessful download: status %s", resp.Status)
					}
					fmt.Printf("Downloaded new binary.\n")
					defer resp.Body.Close()
					err = selfupdate.Apply(resp.Body, selfupdate.Options{})
					if err != nil {
						return fmt.Errorf("apply update: %w", err)
					}
					fmt.Printf("Successfully applied the update.\n")
					return nil
				},
			},
			{
				Name:  "which",
				Usage: "print the detected batch system types",
				Action: func(cCtx *cli.Context) error {
					fmt.Println(batchsystem.DetectBatchSystem())
					return nil
				},
			},
			{
				Name:  "show",
				Usage: "show the job data",
				Flags: []cli.Flag{},
				Action: func(cCtx *cli.Context) error {
					if bs == nil {
						return errUnsupported
					}
					jobName := cCtx.Args().Get(0)

					if err := robin.Show(bs, jobName); err != nil {
						return fmt.Errorf("show error: %w", err)
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
		fmt.Fprintf(os.Stderr, "robin: %v\n", err)
		os.Exit(1)
	}
}
