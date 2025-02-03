package nest

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

const (
	tokenLength      int    = 24
	tokenNonceLength int    = 8
	fallbackPort     uint64 = 2222
	banner           string = `     ━━━━━━━┏┓━━━━━━━━
     ━━━━━━━┃┃━━━━━━━━
     ┏━┓┏━━┓┃┗━┓┏┓┏━━┓
     ┃┏┛┃┏┓┃┃┏┓┃┣┫┃┏┓┃
     ┃┃━┃┗┛┃┃┗┛┃┃┃┃┃┃┃
     ┗┛━┗━━┛┗━━┛┗┛┗┛┗┛
     ━━━━━━━━━━━━━━━━━
     ━━━━━━━━━━━━━━━━━
GREETINGS, PROFESSOR FALKEN.
`
)

func UserPort() uint64 {
	uid := os.Getuid()
	h := sha256.New()
	h.Write([]byte("robin nest uid to port"))
	binary.Write(h, binary.BigEndian, int64(uid))
	hash := h.Sum(nil)
	reader := bytes.NewReader(hash)
	number, err := binary.ReadUvarint(reader)
	if err != nil {
		return fallbackPort
	}
	// Keep the port in the safe range
	if number < 10240 {
		number += 10240
	}
	if number >= 65535 {
		number = number % 65535
	}
	return number
}

func generateRandomString(length int) string {
	rb := make([]byte, length)
	_, err := rand.Read(rb)

	if err != nil {
		panic(err)
	}

	rs := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(rb)
	return rs
}

func writeTokenFile() (token, filename string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("get current working directory: %w", err)
	}
	filename = filepath.Join(cwd, ".robin-token."+generateRandomString(tokenNonceLength))
	token = generateRandomString(tokenLength)
	err = os.WriteFile(filename, []byte(token), 0600)
	if err != nil {
		return "", "", fmt.Errorf("writing token file: %w", err)
	}
	return
}

func kiHandler(ctx ssh.Context, challenger gossh.KeyboardInteractiveChallenge) bool {
	token, tokenFilename, err := writeTokenFile()
	if err != nil {
		log.Printf("write token to file: %v", err)
		return false
	}
	defer os.Remove(tokenFilename)

	question := tokenFilename + "\n"
	answers, err := challenger("", "", []string{question}, []bool{false})
	if err != nil {
		return false
	}
	inputToken := answers[0]
	return inputToken == token
}

func getShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "/bin/sh"
}

func sessionHandlerPty(s ssh.Session) {
	io.WriteString(s, banner)

	ptyReq, winCh, _ := s.Pty()

	cmd := exec.Command(getShell())
	cmd.Env = os.Environ()

	cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
	f, err := pty.Start(cmd)
	if err != nil {
		log.Printf("start pty command: %v", err)
		return
	}
	defer f.Close()

	go func() {
		for win := range winCh {
			setWinsize(f, win.Width, win.Height)
		}
	}()

	go func() {
		io.Copy(f, s)
	}()
	go func() {
		io.Copy(s, f)
	}()
	cmd.Wait()
	s.Exit(cmd.ProcessState.ExitCode())
}

func sessionHandlerSimple(s ssh.Session) {
	cmd := exec.Command(getShell(), []string{"-c", strings.Join(s.Command(), " ")}...)

	cmd.Env = os.Environ()

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()

	err := cmd.Start()
	if err != nil {
		log.Printf("start command: %v", err)
		return
	}

	go func() {
		io.Copy(stdin, s)
	}()
	go func() {
		io.Copy(s, stdout)
	}()
	cmd.Wait()
	s.Exit(cmd.ProcessState.ExitCode())
}

func sessionHandler(s ssh.Session) {
	defer s.Close()
	_, _, isPty := s.Pty()
	if isPty {
		sessionHandlerPty(s)
		return
	}
	sessionHandlerSimple(s)
}

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func kiClientHandler(name, instruction string, questions []string, echos []bool) (answers []string, err error) {
	tokenFilename := strings.TrimSpace(questions[0])
	token, err := os.ReadFile(tokenFilename)
	if err != nil {
		return nil, fmt.Errorf("read token file: %w", err)
	}
	return []string{string(token)}, nil
}

func anotherRobinNestIsRunning() bool {
	config := &gossh.ClientConfig{
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Auth: []gossh.AuthMethod{
			gossh.KeyboardInteractive(kiClientHandler),
		},
	}
	addr := "localhost:" + strconv.FormatUint(UserPort(), 10)
	conn, err := gossh.Dial("tcp", addr, config)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func RunShellServer(nodeHead chan struct{}) error {
	if anotherRobinNestIsRunning() {
		return nil
	}
	userPort := UserPort()
	server := &ssh.Server{
		Addr:                       ":" + strconv.FormatUint(userPort, 10),
		Handler:                    sessionHandler,
		KeyboardInteractiveHandler: kiHandler,
	}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("shell listen: %w", err)
	}

	// Mark this instance as the node head
	close(nodeHead)

	err = server.Serve(listener)
	if err != nil {
		return fmt.Errorf("serve ssh: %w", err)
	}
	return nil
}
