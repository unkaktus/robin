package slurm

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/unkaktus/spanner"
)

type ListedJob struct {
	Name             string
	ID               string
	Partition        string
	State            string
	ExitCode         string
	SubmitTime       string
	NodeList         string
	NodeNumber       string
	TimeUsed         string
	TimeLimit        string
	OutputFile       string
	ErrorFile        string
	WorkingDirectory string
}

const (
	outlen = 16384
)

func SqueueRequestString(fields []string) string {
	ret := ""
	for _, field := range fields {
		ret = fmt.Sprintf("%s,%s:%d", ret, field, outlen)
	}
	return ret
}

func valueString(data []byte, i int) string {
	return strings.TrimRight(string(data[i:i+outlen]), " ")
}

func UnmarshalSqueueOutput(data []byte) (ListedJob, error) {
	listedJob := ListedJob{}

	i := 0
	listedJob.Name = valueString(data, i)
	i += outlen
	listedJob.ID = valueString(data, i)
	i += outlen
	listedJob.Partition = valueString(data, i)
	i += outlen
	listedJob.State = valueString(data, i)
	i += outlen
	listedJob.ExitCode = valueString(data, i)
	i += outlen
	listedJob.SubmitTime = valueString(data, i)
	i += outlen
	listedJob.NodeList = valueString(data, i)
	i += outlen
	listedJob.NodeNumber = valueString(data, i)
	i += outlen
	listedJob.TimeUsed = valueString(data, i)
	i += outlen
	listedJob.TimeLimit = valueString(data, i)
	i += outlen
	listedJob.OutputFile = valueString(data, i)
	i += outlen
	listedJob.ErrorFile = valueString(data, i)
	i += outlen
	listedJob.WorkingDirectory = valueString(data, i)

	return listedJob, nil
}

func query(all bool) ([]ListedJob, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}
	req := []string{
		"Name",
		"JobID",
		"Partition",
		"StateCompact",
		"exit_code",
		"SubmitTime",
		"NodeList",
		"NumNodes",
		"TimeUsed",
		"TimeLimit",
		"STDOUT",
		"STDERR",
		"WorkDir",
	}
	squeueArguments := []string{
		"--noheader",
		"-O", SqueueRequestString(req),
	}
	if !all {
		squeueArguments = append(squeueArguments,
			[]string{"-u", currentUser.Uid}...,
		)
	}
	cmd := exec.Command("squeue", squeueArguments...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("exectute command: %w", err)
	}

	listedJobs := []ListedJob{}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	scanner.Buffer(nil, 16*outlen)
	for scanner.Scan() {
		listedJob, err := UnmarshalSqueueOutput(scanner.Bytes())
		if err != nil {
			return nil, fmt.Errorf("unmarshal squeue output:: %w", err)
		}
		listedJobs = append(listedJobs, listedJob)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return listedJobs, nil
}

func expandRangeString(rangeString string) []string {
	ret := []string{}
	commaSplit := strings.Split(rangeString, ",")
	for _, cs := range commaSplit {
		dashSplit := strings.Split(cs, "-")
		digits := len(dashSplit[0])
		left, _ := strconv.Atoi(dashSplit[0])
		right := left
		if len(dashSplit) == 2 {
			right, _ = strconv.Atoi(dashSplit[1])
		}
		for i := left; i <= right; i++ {
			s := fmt.Sprintf("%0"+strconv.Itoa(digits)+"d", i)
			ret = append(ret, s)
		}
	}
	return ret
}

func parseNodeList(nodelistString string) ([]string, error) {
	nodelist := []string{}
	if nodelistString == "" {
		return nodelist, nil
	}

	prefix := ""
	suffix := ""
	rangeString := ""
	insideRange := false
	writeSuffix := false
	for _, r := range nodelistString {
		switch r {
		case '[':
			insideRange = true
		case ']':
			insideRange = false
			writeSuffix = true
		case ',':
			if insideRange {
				rangeString += string(r)
			}
			if !insideRange {
				if rangeString == "" {
					node := prefix
					nodelist = append(nodelist, node)
					prefix = ""
					continue
				}
				// Finalize bunch
				expandedRange := expandRangeString(rangeString)
				for _, ers := range expandedRange {
					node := prefix + ers + suffix

					nodelist = append(nodelist, node)
				}
				prefix = ""
				rangeString = ""
				suffix = ""
				writeSuffix = false
			}
		default:
			if !insideRange {
				switch writeSuffix {
				case false:
					prefix += string(r)
				case true:
					suffix += string(r)
				}
			} else {
				rangeString += string(r)
			}
		}
	}
	// Finalize last bunch
	if rangeString == "" {
		node := prefix
		nodelist = append(nodelist, node)
	} else {
		expandedRange := expandRangeString(rangeString)
		for _, ers := range expandedRange {
			node := prefix + ers + suffix
			nodelist = append(nodelist, node)
		}
	}

	return nodelist, nil
}

const (
	TimeLayout = "2006-01-02T15:04:05"
)

func clockDuration(clock string) (d time.Duration, err error) {
	days := 0
	daysp := strings.Split(clock, "-")
	if len(daysp) == 2 {
		clock = daysp[1]
		days, err = strconv.Atoi(daysp[0])
		if err != nil {
			return 0, fmt.Errorf("parse day component: %w", err)
		}
	}

	sp := strings.Split(clock, ":")
	switch len(sp) {
	case 3:
		h, m, s := sp[0], sp[1], sp[2]
		d, err = time.ParseDuration(fmt.Sprintf("%sh%sm%ss", h, m, s))
		if err != nil {
			return 0, fmt.Errorf("parse duration: %w", err)
		}
	case 2:
		m, s := sp[0], sp[1]
		d, err = time.ParseDuration(fmt.Sprintf("%sm%ss", m, s))
		if err != nil {
			return 0, fmt.Errorf("parse duration: %w", err)
		}
	default:
		return 0, fmt.Errorf("wrong string length")
	}

	d += time.Duration(days) * 24 * time.Hour
	return d, nil
}

func parseExitCode(s string) (exitCode, signal int, err error) {
	sp := strings.Split(s, ":")
	exitCode, err = strconv.Atoi(sp[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse exit code: %w", err)
	}
	if len(sp) > 1 {
		signal, err = strconv.Atoi(sp[1])
		if err != nil {
			return 0, 0, fmt.Errorf("parse signal: %w", err)
		}
	}
	return exitCode, signal, nil
}

func listOutputToJobList(listedJobs []ListedJob) (jobs []spanner.Job, err error) {
	for _, listedJob := range listedJobs {
		exitCode, _, err := parseExitCode(listedJob.ExitCode)
		if err != nil {
			return nil, fmt.Errorf("parse exit code: %w", err)
		}
		creationTime, err := time.Parse(TimeLayout, listedJob.SubmitTime)
		if err != nil {
			return nil, fmt.Errorf("parse submit time: %w", err)
		}
		nodes, err := parseNodeList(listedJob.NodeList)
		if err != nil {
			return nil, fmt.Errorf("parse nodelist: %w", err)
		}
		nodeNumber, err := strconv.Atoi(listedJob.NodeNumber)
		if err != nil {
			return nil, fmt.Errorf("parse node number: %w", err)
		}
		walltime, err := clockDuration(listedJob.TimeUsed)
		if err != nil {
			return nil, fmt.Errorf("parse walltime: %w", err)
		}
		requestedWalltime := time.Duration(0)
		if listedJob.TimeLimit != "UNLIMITED" {
			requestedWalltime, err = clockDuration(listedJob.TimeLimit)
			if err != nil {
				return nil, fmt.Errorf("parse requested walltime: %w", err)
			}
		}
		job := spanner.Job{
			Name:              listedJob.Name,
			ID:                listedJob.ID,
			Queue:             listedJob.Partition,
			State:             listedJob.State,
			ExitCode:          exitCode,
			CreationTime:      creationTime,
			Nodes:             nodes,
			NodeNumber:        nodeNumber,
			Walltime:          walltime,
			RequestedWalltime: requestedWalltime,
			OutputFile:        listedJob.OutputFile,
			ErrorFile:         listedJob.ErrorFile,
			WorkingDirectory:  listedJob.WorkingDirectory,
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (b *Slurm) ListJobs(all bool) ([]spanner.Job, error) {
	listOutput, err := query(all)
	if err != nil {
		return nil, fmt.Errorf("query list: %w", err)
	}

	jobList, err := listOutputToJobList(listOutput)
	if err != nil {
		return nil, fmt.Errorf("convert to job list: %w", err)
	}

	return jobList, nil
}

func (b *Slurm) FindJob(jobName string) (*spanner.Job, error) {
	jobList, err := b.ListJobs(false)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	for _, job := range jobList {
		if job.Name == jobName {
			return &job, nil
		}
	}
	return nil, nil
}
