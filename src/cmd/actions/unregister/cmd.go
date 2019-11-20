package unregister

import (
	"denver/cmd"
	"denver/pkg/providers"
	"fmt"
	"log"
)

// Unregister action
type Unregister struct {
	vmProvider *providers.VMProvider
	printer    *log.Logger
}

// NewUnregister returns a pointer to Unregister
func NewUnregister(vmProvider *providers.VMProvider, printer *log.Logger) *Unregister {
	return &Unregister{
		vmProvider: vmProvider,
		printer:    printer,
	}
}

// GetCommand returns a valid cmd command
func (d *Unregister) GetCommand() cmd.DenverCommand {
	return cmd.DenverCommand{
		Name: "unregister",
		Desc: "Unregister the instance from the provider",
		Exec: func() error {
			state := (*d.vmProvider).GetState()
			if state.Live {
				return fmt.Errorf("VM is started, stop it first")
			}

			return (*d.vmProvider).Unregister()
		},
	}
}
