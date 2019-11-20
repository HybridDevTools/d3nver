package checkversion

import (
	"denver/cmd"
	"denver/pkg/notify"
	"denver/pkg/providers"
	"denver/pkg/updater"
	"fmt"
	"log"

	"github.com/logrusorgru/aurora"
)

// CheckVersion action
type CheckVersion struct {
	vmProvider *providers.VMProvider
	updater    updater.Updater
	printer    *log.Logger
	notify     notify.Notify
}

// NewCheckVersion returns a pointer to CheckVersion
func NewCheckVersion(vmProvider *providers.VMProvider, updater updater.Updater, printer *log.Logger, notify notify.Notify) *CheckVersion {
	return &CheckVersion{
		vmProvider: vmProvider,
		updater:    updater,
		printer:    printer,
		notify:     notify,
	}
}

// GetCommand returns a valid cmd command
func (c *CheckVersion) GetCommand() cmd.DenverCommand {
	return cmd.DenverCommand{
		Name: "check-version",
		Desc: "Checks if you are running latest Denver version",
		Exec: func() (err error) {
			controllerUpToDate, err := c.checkForControllerUpdate()
			if err != nil {
				return
			}

			rbiUpToDate, err := c.checkForRootBaseImageUpdate()
			if err != nil {
				return
			}

			if controllerUpToDate && rbiUpToDate {
				c.printer.Println(fmt.Sprintf("%s %s",
					aurora.Bold(aurora.Green("[OK]")),
					"Your versions are up to date",
				))
			}

			return nil
		},
	}
}

// CheckForUpdates checks whether the Controller or the Root Base Image has new updates
func (c *CheckVersion) CheckForUpdates() (err error) {
	if _, err = c.checkForControllerUpdate(); err != nil {
		return
	}
	if _, err = c.checkForRootBaseImageUpdate(); err != nil {
		return
	}
	return
}

func (c *CheckVersion) checkForControllerUpdate() (isUpToDate bool, err error) {
	_, isUpToDate, err = c.updater.CheckIsUpdated(cmd.BuildTs)
	if err != nil {
		return
	}
	if isUpToDate {
		return
	}

	answer := c.notify.AskQuestion("A new version for Denver is available, do you want to update?")
	if !answer {
		return
	}

	err = c.updater.Update()
	return
}

func (c *CheckVersion) checkForRootBaseImageUpdate() (isUpToDate bool, err error) {
	isUpToDate, err = (*c.vmProvider).CheckIsUpdated()
	if err != nil {
		return
	}
	if isUpToDate {
		return
	}

	answer := c.notify.AskQuestion("A new version for the Root Base Image is available, do you want to update?")
	if !answer {
		return
	}

	_, err = (*c.vmProvider).Update()
	return
}
