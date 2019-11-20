package ssh

import (
	"os"

	"github.com/docker/docker/pkg/term"
	"golang.org/x/crypto/ssh"
)

type terminal struct {
	client *ssh.Client
}

// Terminal opens a new Terminal against an SSL connection
func (s *SSH) Terminal() (err error) {
	client, err := s.connect()
	if err != nil {
		return
	}
	defer client.Close()

	return (&terminal{client: client}).Terminal()
}

func (t *terminal) Terminal() (err error) {
	var (
		termWidth, termHeight int
	)

	session, err := t.client.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO: 1,
	}

	fd := os.Stdin.Fd()

	if term.IsTerminal(fd) {
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			return err
		}

		defer func() {
			_ = term.RestoreTerminal(fd, oldState)
		}()

		winSize, err := term.GetWinsize(fd)
		if err != nil {
			termWidth = 80
			termHeight = 24
		} else {
			termWidth = int(winSize.Width)
			termHeight = int(winSize.Height)
		}
	}

	if err = session.RequestPty("xterm", termHeight, termWidth, modes); err != nil {
		return
	}

	if err = session.Shell(); err != nil {
		return
	}
	if err = session.Wait(); err != nil {
		return
	}

	return
}
