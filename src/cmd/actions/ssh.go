package actions

import (
	"denver/cmd"
	"denver/pkg/providers"
	"denver/pkg/ssh"
	"fmt"
	"log"
)

// SSH action
type SSH struct {
	ssh        *ssh.SSH
	vmProvider *providers.VMProvider
	printer    *log.Logger
}

// NewSSH returns a pointer to SSH
func NewSSH(ssh *ssh.SSH, vmProvider *providers.VMProvider, printer *log.Logger) *SSH {
	return &SSH{
		ssh:        ssh,
		vmProvider: vmProvider,
		printer:    printer,
	}
}

// GetCommand returns a valid cmd command
func (s *SSH) GetCommand() cmd.DenverCommand {
	return cmd.DenverCommand{
		Name: "ssh",
		Desc: "Connect through ssh in local terminal",
		Exec: func() error {
			state := (*s.vmProvider).GetState()
			if !state.AllSystemsReady {
				return fmt.Errorf("VM not ready")
			}

			return s.ssh.Terminal()
		},
	}
}
