package tent

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
)

type Variables struct {
	ConfigFilename  string
	TaskID          int
	TotalTaskNumber int
}

func RunCommand(cmdline []string, vars Variables) (process *os.Process, err error) {
	for i, arg := range cmdline {
		cmdline[i], err = ExecTemplate(arg, vars)
		if err != nil {
			return nil, fmt.Errorf("executing template on argument %s: %w", arg, err)
		}
	}
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("task start: %v", err)
	}
	return cmd.Process, nil
}

func ExecTemplate(ts string, s interface{}) (string, error) {
	t, err := template.New("template").Parse(ts)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)

	}
	builder := &strings.Builder{}

	err = t.Execute(builder, s)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return builder.String(), nil
}

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func RunShellServer() error {
	ssh.Handle(func(s ssh.Session) {
		io.WriteString(s, "Connected to spanner shell.\n")
		cmd := exec.Command("bash")
		ptyReq, winCh, isPty := s.Pty()
		if isPty {
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
		} else {
			io.WriteString(s, "No PTY requested.\n")
			s.Exit(1)
		}
	})

	err := ssh.ListenAndServe(":2222", nil)
	if err != nil {
		return fmt.Errorf("listen and serve ssh: %w", err)
	}
	return nil
}
