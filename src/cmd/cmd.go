package cmd

import (
	"github.com/spf13/cobra"
)

// Version compiled
var Version = "dev"

// BuildTs stores ts of compilation
var BuildTs = "0"

// WorkingDirectory will be unset during compilation
var WorkingDirectory = "."

// Action interface must be implemented to define a new CLI action
type Action interface {
	GetCommand() DenverCommand
}

// DenverCommand contains a CLI command
type DenverCommand struct {
	Name string
	Desc string
	Exec func() error
}

// CreateCobraCommand returns a a cobra command from an DenverCommand
func CreateCobraCommand(command DenverCommand) *cobra.Command {
	return &cobra.Command{
		Use:   command.Name,
		Short: command.Desc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return command.Exec()
		},
	}
}
