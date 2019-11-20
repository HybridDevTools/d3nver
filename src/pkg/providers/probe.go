package providers

import (
	"context"
	"denver/pkg/ssh"
	"log"
	"strings"
	"time"
)

// Probe struct allow us to get VM status
type Probe struct {
	ssh    *ssh.SSH
	ticker *time.Ticker
	ctx    context.Context
}

// NewProbe returns a pointer to Probe
func NewProbe(ctx context.Context, ssh *ssh.SSH) *Probe {
	return &Probe{
		ssh: ssh,
		ctx: ctx,
	}
}

// Start polling a VM instance
func (s *Probe) Start(vmProvider VMProvider) error {
	s.ticker = time.NewTicker(250 * time.Millisecond)

	if err := s.probe(vmProvider); err != nil {
		return err
	}

	go func() {
		defer s.ticker.Stop()

		for {
			select {
			case <-s.ticker.C:
				_ = s.probe(vmProvider)
			case <-s.ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (s *Probe) probe(vmProvider VMProvider) (err error) {
	vmState := NewState()
	defer func() {
		err := vmProvider.setState(vmState)
		if err != nil {
			log.Println(err)
		}
	}()

	// Check is VM is alive
	if vmState.Live, err = vmProvider.checkIfRunning(); err != nil {
		return
	}

	if vmState.Live != true {
		return
	}

	// Check if all systems are ready on the Virtual Machine
	if err = s.checkAllSystemsReady(vmState); err != nil {
		return
	}

	return
}

func (s *Probe) checkAllSystemsReady(VMState *State) (err error) {
	out, err := s.ssh.Cmd("echo 'OK'")
	VMState.OsReady = err == nil
	if err != nil {
		return
	}

	cmdReturn := strings.TrimSpace(out)
	VMState.AllSystemsReady = cmdReturn == "OK"

	return
}
