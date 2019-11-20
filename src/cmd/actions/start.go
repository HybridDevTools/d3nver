package actions

import (
	"context"
	"denver/cmd"
	"denver/cmd/actions/checkversion"
	"denver/pkg/providers"
	"fmt"
	"log"
	"time"

	"github.com/logrusorgru/aurora"
)

// Start action
type Start struct {
	vmProvider   *providers.VMProvider
	printer      *log.Logger
	ctx          context.Context
	checkVersion *checkversion.CheckVersion
}

// NewStart returns a pointer to Start
func NewStart(ctx context.Context, vmProvider *providers.VMProvider, printer *log.Logger, checkVersion *checkversion.CheckVersion) *Start {
	return &Start{
		vmProvider:   vmProvider,
		printer:      printer,
		ctx:          ctx,
		checkVersion: checkVersion,
	}
}

// GetCommand returns a valid cmd command
func (s *Start) GetCommand() cmd.DenverCommand {
	return cmd.DenverCommand{
		Name: "start",
		Desc: "Start the instance",
		Exec: func() (err error) {
			if err = (*s.checkVersion).CheckForUpdates(); err != nil {
				return
			}

			if err = (*s.vmProvider).Start(); err != nil {
				return
			}

			s.printer.Print(fmt.Sprintf("%s %s",
				aurora.Bold(aurora.Yellow("[INFO]")),
				"VM is starting...",
			))

			tick := time.Tick(250 * time.Millisecond)
			state := (*s.vmProvider).GetState()

			for !state.AllSystemsReady {
				select {
				case <-tick:
					state = (*s.vmProvider).GetState()
				case <-s.ctx.Done():
					return
				}
			}

			s.printer.Println(fmt.Sprintf("%s %s",
				aurora.Bold(aurora.Green("[OK]")),
				"VM has been started",
			))

			return
		},
	}
}
