package ssh

import (
	"bytes"

	"golang.org/x/crypto/ssh"
)

type cmd struct {
	client *ssh.Client
}

// Cmd executes a command over an SSH connection
func (s *SSH) Cmd(command string) (out string, err error) {
	client, err := s.connect()
	if err != nil {
		return
	}
	defer client.Close()

	return (&cmd{client: client}).cmd(command)
}

func (c *cmd) cmd(cmd string) (out string, err error) {
	session, err := c.client.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(cmd)
	out = stdout.String()
	if err != nil {
		out = stderr.String()
	}

	return
}
