package begin

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
)

type Beginfile struct {
	Name string

	Nodes        int
	TasksPerNode int

	NodeType string
	Walltime time.Duration
	Email    string

	WorkingDirectory string

	InitScript []string
	Runtime    []string
	Executable string
	Arguments  []string

	PostScript []string
}

func ParseBeginfile(filename string) (*Beginfile, error) {
	beginfile := &Beginfile{}

	_, err := toml.DecodeFile(filename, beginfile)
	if err != nil {
		return nil, fmt.Errorf("decode file: %w", err)
	}

	return beginfile, nil
}
