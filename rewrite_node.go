package robin

import (
	"os/exec"
	"strings"
)

func isSupermuc() bool {
	cmd := exec.Command("hostname", "-d")
	combi, _ := cmd.CombinedOutput()
	return strings.Contains(string(combi), "sng.lrz.de")
}

func RewriteNode(node string) string {
	switch {
	case isSupermuc():
		return node + "opa"
	}
	return node
}
