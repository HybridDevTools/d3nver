package actions

import (
	"denver/cmd"
	"denver/pkg/providers"
	"fmt"
	"log"

	"github.com/logrusorgru/aurora"
)

// Status action
type Status struct {
	vmProvider *providers.VMProvider
	printer    *log.Logger
}

// NewStatus returns a pointer to Status
func NewStatus(vmProvider *providers.VMProvider, printer *log.Logger) *Status {
	return &Status{
		vmProvider: vmProvider,
		printer:    printer,
	}
}

// GetCommand returns a valid cmd command
func (s *Status) GetCommand() cmd.DenverCommand {
	return cmd.DenverCommand{
		Name: "status",
		Desc: "Check if the instance is ready to operate",
		Exec: func() error {
			state := (*s.vmProvider).GetState()

			var message string
			if state.Live == true {
				message = fmt.Sprintf("%s %s",
					aurora.Bold(aurora.Green("[OK]")),
					"Virtual machine state is power on",
				)
			} else {
				message = fmt.Sprintf("%s %s",
					aurora.Bold(aurora.Red("[KO]")),
					"Virtual machine state is power off",
				)
			}
			s.printer.Println(message)

			if state.OsReady == true {
				message = fmt.Sprintf("%s %s",
					aurora.Bold(aurora.Green("[OK]")),
					"Virtual machine OS is ready to handle with SSH",
				)
			} else {
				message = fmt.Sprintf("%s %s",
					aurora.Bold(aurora.Red("[KO]")),
					"SSH on Virtual machine OS isn't ready",
				)
			}
			s.printer.Println(message)

			if state.AllSystemsReady == true {
				message = fmt.Sprintf("%s %s",
					aurora.Bold(aurora.Green("[OK]")),
					"Virtual machine systems are up",
				)
			} else {
				message = fmt.Sprintf("%s %s",
					aurora.Bold(aurora.Red("[KO]")),
					"At least one of the Virtual machine systems is down",
				)
			}
			s.printer.Println(message)

			return nil
		},
	}
}
