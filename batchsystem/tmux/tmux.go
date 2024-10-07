package tmux

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/unkaktus/robin"
)

type Tmux struct {
}

func (tm *Tmux) Init() error {
	if err := exec.Command("tmux", "new-session", "-d", "tmux set-option exit-empty off").Run(); err != nil {
		return fmt.Errorf("start tmux server: %w", err)
	}
	return nil
}

type NameData struct {
	robin.Comment
	Name    string `json:"name"`
	LogFile string `json:"log_file"`
}

func (nd *NameData) DecodeString(s string) error {
	nameDataEndcoded := strings.TrimPrefix(s, "robin_")
	nameDataDecoded, err := base64.RawURLEncoding.DecodeString(nameDataEndcoded)
	if err != nil {
		return fmt.Errorf("decode base64: %w", err)
	}
	err = json.Unmarshal(nameDataDecoded, nd)
	if err != nil {
		return fmt.Errorf("unmarshal JSON: %w", err)
	}
	return nil
}

func (nd NameData) EncodeToString() string {
	nameDataJSON, _ := json.Marshal(nd)
	nameDataEncoded := "robin_" + base64.RawURLEncoding.EncodeToString(nameDataJSON)
	return nameDataEncoded
}
