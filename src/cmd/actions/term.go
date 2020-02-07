package actions

import (
	"denver/cmd"
	"denver/pkg/providers"
	"denver/pkg/util/executor"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
)

// Term action
type Term struct {
	workingDirectory                      string
	executor                              *executor.Executor
	user, ip, terminal, terminalArguments *string
	vmProvider                            *providers.VMProvider
	printer                               *log.Logger
}

// This variables has been created to be set during compilation :)
var windowsTerm = "alacritty-windows-0.4.1.exe"
var darwinTerm = "alacritty-darwin-0.4.1"
var linuxTerm = "alacritty-linux-0.4.1"

// NewTerm returns a pointer to Term
func NewTerm(workingDirectory string, user, ip, terminal, terminalArguments *string, vmProvider *providers.VMProvider, printer *log.Logger) *Term {
	return &Term{
		workingDirectory:  workingDirectory,
		executor:          executor.NewExecutor(),
		user:              user,
		ip:                ip,
		terminal:          terminal,
		terminalArguments: terminalArguments,
		vmProvider:        vmProvider,
		printer:           printer,
	}
}

// GetCommand returns a valid cmd command
func (t *Term) GetCommand() cmd.DenverCommand {
	return cmd.DenverCommand{
		Name: "term",
		Desc: "Connect through configured terminal",
		Exec: func() error {
			state := (*t.vmProvider).GetState()
			if !state.AllSystemsReady {
				return fmt.Errorf("VM not ready")
			}

			var command []string
			var arguments []string
			var defaultTerm string
			user := fmt.Sprintf("%s@%s", *t.user, *t.ip)
			terminal := *t.terminal
			tArguments := *t.terminalArguments
			if terminal == "default" {
				switch runtime.GOOS {
				case "windows":
					defaultTerm = windowsTerm
				case "darwin":
					defaultTerm = darwinTerm
				case "linux":
					defaultTerm = linuxTerm
				}
				terminal = filepath.Join(t.workingDirectory, "tools", defaultTerm)
				arguments = append(arguments,
					"--config-file", filepath.Join(t.workingDirectory, "tools", "alacritty.yml"),
					"--working-directory", t.workingDirectory,
					"-e",
				)
			} else if terminal == "iterm2" {
				switch runtime.GOOS {
				case "windows":
				case "linux":
					return fmt.Errorf("not supported on this system")
				case "darwin":
					defaultTerm = "iterm2.sh"
				}
				terminal = filepath.Join(t.workingDirectory, "tools", defaultTerm)
			}

			command = append(command, terminal)
			if tArguments != "" {
				arguments = append(arguments, tArguments)
			}
			arguments = append(arguments, "ssh", "-i", ".ssh/id_rsa", "-o", "StrictHostKeyChecking=no", user)
			command = append(command, arguments...)
			if _, err := t.executor.Execute(command); err != nil {
				return err
			}

			return nil
		},
	}
}
