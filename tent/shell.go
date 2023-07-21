package tent

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

const (
	TokenLength int = 16
)

func generateToken() string {
	rb := make([]byte, TokenLength)
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
	filename = filepath.Join(cwd, ".spanner-token")
	token = generateToken()
	err = ioutil.WriteFile(filename, []byte(token), 0600)
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

func sessionHandler(s ssh.Session) {
	io.WriteString(s, "Connected to spanner shell.\n")
	cmd := exec.Command("bash")
	ptyReq, winCh, isPty := s.Pty()
	if !isPty {
		io.WriteString(s, "No PTY requested.\n")
		s.Exit(1)
		return
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
	f, err := pty.Start(cmd)
	if err != nil {
		panic(err)
	}
	go func() {
		for win := range winCh {
			setWinsize(f, win.Width, win.Height)
		}
	}()
	go func() {
		io.Copy(f, s) // stdin
	}()
	io.Copy(s, f) // stdout
	cmd.Wait()
}

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func RunShellServer() error {
	server := &ssh.Server{
		Addr:                       ":2222",
		Handler:                    sessionHandler,
		KeyboardInteractiveHandler: kiHandler,
	}

	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("listen and serve ssh: %w", err)
	}
	return nil
}
