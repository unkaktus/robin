package spanner

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func kiHandler(name, instruction string, questions []string, echos []bool) (answers []string, err error) {
	tokenFilename := strings.TrimSpace(questions[0])
	token, err := ioutil.ReadFile(tokenFilename)
	if err != nil {
		return nil, fmt.Errorf("read token file: %w", err)
	}
	return []string{string(token)}, nil
}

func Shell(hostname string) error {
	config := &gossh.ClientConfig{
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Auth: []gossh.AuthMethod{
			gossh.KeyboardInteractive(kiHandler),
		},
	}
	conn, err := gossh.Dial("tcp", hostname+":2222", config)
	if err != nil {
		return fmt.Errorf("dial ssh: %w", err)
	}
	defer conn.Close()
	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	defer session.Close()

	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("terminal make raw: %s", err)
	}
	defer term.Restore(fd, state)

	w, h, err := term.GetSize(fd)
	if err != nil {
		return fmt.Errorf("terminal get size: %s", err)
	}
	modes := gossh.TerminalModes{
		gossh.ECHO:          1,
		gossh.TTY_OP_ISPEED: 14400,
		gossh.TTY_OP_OSPEED: 14400,
	}
	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm-256color"
	}
	if err := session.RequestPty(term, h, w, modes); err != nil {
		return fmt.Errorf("session xterm: %s", err)
	}

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err = session.Start("")
	if err != nil {
		return fmt.Errorf("start session: %w", err)
	}
	return session.Wait()
}
