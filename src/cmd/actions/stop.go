package actions

import (
	"context"
	"denver/cmd"
	"denver/pkg/providers"
	"fmt"
	"log"
	"time"

	"github.com/logrusorgru/aurora"
)

// Stop action
type Stop struct {
	vmProvider *providers.VMProvider
	printer    *log.Logger
	ctx        context.Context
}

// NewStop returns a pointer to Stop
func NewStop(ctx context.Context, vmProvider *providers.VMProvider, printer *log.Logger) *Stop {
	return &Stop{
		vmProvider: vmProvider,
		printer:    printer,
		ctx:        ctx,
	}
}

// GetCommand returns a valid cmd command
func (s *Stop) GetCommand() cmd.DenverCommand {
	return cmd.DenverCommand{
		Name: "stop",
		Desc: "Stop the instance",
		Exec: func() (err error) {
			state := (*s.vmProvider).GetState()
			if !state.Live {
				s.printer.Println(fmt.Sprintf("%s %s",
					aurora.Bold(aurora.Yellow("[SKIP]")),
					"VM already stopped",
				))
				return
			}

			if err = (*s.vmProvider).Stop(); err != nil {
				return err
			}

			s.printer.Print(fmt.Sprintf("%s %s",
				aurora.Bold(aurora.Yellow("[INFO]")),
				"VM is stopping...",
			))
			tick := time.Tick(250 * time.Millisecond)
			for state.Live {
				select {
				case <-tick:
					state = (*s.vmProvider).GetState()
				case <-s.ctx.Done():
					return
				}
			}

			s.printer.Println(fmt.Sprintf("%s %s",
				aurora.Bold(aurora.Green("[OK]")),
				"VM has been stopped",
			))

			return
		},
	}
}
