package actions

import (
	"denver/cmd"
	"denver/cmd/actions/checkversion"
	"denver/pkg/providers"
	"fmt"
	"log"

	"github.com/logrusorgru/aurora"
)

// Init action
type Init struct {
	vmProvider   *providers.VMProvider
	printer      *log.Logger
	checkVersion *checkversion.CheckVersion
}

// NewInit returns a pointer to Init
func NewInit(vmProvider *providers.VMProvider, printer *log.Logger, checkVersion *checkversion.CheckVersion) *Init {
	return &Init{
		vmProvider:   vmProvider,
		printer:      printer,
		checkVersion: checkVersion,
	}
}

// GetCommand returns a valid cmd command
func (i *Init) GetCommand() cmd.DenverCommand {
	return cmd.DenverCommand{
		Name: "init",
		Desc: "Init the instance",
		Exec: func() (err error) {
			if err = (*i.checkVersion).CheckForUpdates(); err != nil {
				return
			}

			if err = (*i.vmProvider).Init(); err != nil {
				return
			}

			i.printer.Println(fmt.Sprintf("%s %s",
				aurora.Bold(aurora.Green("[OK]")),
				"VM has been installed",
			))

			return
		},
	}
}
