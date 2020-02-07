package root

import (
	"context"
	"denver/cmd"
	"denver/cmd/actions"
	"denver/cmd/actions/checkversion"
	"denver/cmd/actions/unregister"
	"denver/pkg/notify"
	"denver/pkg/providers"
	"denver/pkg/ssh"
	"denver/pkg/storage/http"
	"denver/pkg/updater"
	"denver/pkg/user"
	"denver/pkg/util/compressor"
	"denver/structs"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"

	"github.com/logrusorgru/aurora"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Denver holds main configuration for the Denver application
type Denver struct {
	workingDirectory string
	printer          *log.Logger
	config           *structs.Denver
	availableActions []cmd.Action
	bootstrapFunc    []func() error
	vMProvider       providers.VMProvider
	ssh              *ssh.SSH
	updater          updater.Updater
	ctx              context.Context
	notify           notify.Notify
}

var configFile string

// New returns a pointer to Denver
func New(ctx context.Context, workingDirectory string) *Denver {
	setLog()
	return &Denver{
		workingDirectory: workingDirectory,
		printer:          log.New(os.Stdout, "", 0),
		config:           structs.NewDenverConfig(),
		ssh:              &ssh.SSH{},
		updater: updater.NewDefaultUpdater(
			ctx,
			workingDirectory,
			fmt.Sprintf("%s/%s/manifest.json", cmd.UpdatePath, runtime.GOOS),
			[]string{
				"denver",
				"tools",
				filepath.Join("conf", "config.dist.yml"),
			},
			compressor.NewMultiCompressor(),
			&http.HTTP{},
		),
		ctx:    ctx,
		notify: notify.CliQuestion{},
	}
}

// Execute bootstraps main application
func (s *Denver) Execute() int {
	s.addActions()
	s.addBootstrapFunc()

	rootCmd := s.getRootCommand()
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Config file")
	for _, action := range s.availableActions {
		rootCmd.AddCommand(cmd.CreateCobraCommand(action.GetCommand()))
	}

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
		for _, f := range s.bootstrapFunc {
			if err = f(); err != nil {
				return
			}
		}
		return
	}

	if err := rootCmd.Execute(); err != nil {
		s.printer.Println(fmt.Sprintf("%s %s",
			aurora.Bold(aurora.Red("[KO]")),
			err.Error(),
		))
		return 1
	}

	return 0
}

func (s *Denver) getRootCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "denver",
		Short:         "Helper minion for the D3nver platform",
		Long:          "D3nver, the Developer ENVironment",
		Version:       fmt.Sprintf("%s", cmd.Version),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

func (s *Denver) initConfig() (err error) {
	if configFile == "" {
		home, err := homedir.Dir()
		if err != nil {
			return err
		}

		paths := []string{
			".",
			s.workingDirectory,
			home,
		}

		for _, path := range paths {
			viper.AddConfigPath(path)
			viper.AddConfigPath(filepath.Join(path, "conf"))
		}

		viper.SetConfigName("config")
	} else {
		viper.SetConfigFile(configFile)
	}

	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err != nil {
		return
	}

	return viper.Unmarshal(s.config, func(c *mapstructure.DecoderConfig) {
		c.ErrorUnused = true
	})
}

func (s *Denver) setVMProvider() (err error) {
	var provider structs.Provider
	var ok bool
	if provider, ok = s.config.Providers[s.config.Instance.Provider]; !ok {
		return fmt.Errorf("VM Provider %s not found", s.config.Instance.Provider)
	}

	if s.vMProvider, err = providers.GetVMProvider(
		s.ctx,
		provider,
		s.config.Instance,
		s.config.UserInfo.Userdatasize,
		s.config.Config.Channel,
		s.config.Config.RBIURL,
		s.workingDirectory,
	); err != nil {
		return
	}

	u := user.NewUser(s.config.UserInfo, s.ssh)
	s.vMProvider.AddPostStartAction(func() (err error) {
		return u.SetGitUser()
	})

	s.vMProvider.AddPostStartAction(func() (err error) {
		return u.SetUserKey()
	})

	return
}

func (s *Denver) setSSH() (err error) {
	sshVal, err := ssh.NewSSH(s.config.Instance.Localip, s.workingDirectory)
	if err != nil {
		return
	}
	*s.ssh = *sshVal

	return
}

func (s *Denver) initProbe() (err error) {
	return providers.NewProbe(s.ctx, s.ssh).Start(s.vMProvider)
}

func setLog() {
	log.SetOutput(os.Stdout)
	log.SetPrefix("[D3NVER] ")
	log.SetFlags(log.LUTC | log.LstdFlags | log.Lshortfile)
}

func (s *Denver) addBootstrapFunc() {
	s.bootstrapFunc = append(s.bootstrapFunc,
		func() (err error) {
			return s.initConfig()
		},
		func() (err error) {
			return s.setVMProvider()
		},
		func() (err error) {
			return s.setSSH()
		},
		func() (err error) {
			return s.initProbe()
		},
	)
}

func (s *Denver) addActions() {
	checkVersion := checkversion.NewCheckVersion(&s.vMProvider, s.updater, s.printer, s.notify)

	s.availableActions = append(
		s.availableActions,
		actions.NewInit(&s.vMProvider, s.printer, checkVersion),
		actions.NewSSH(s.ssh, &s.vMProvider, s.printer),
		actions.NewStart(s.ctx, &s.vMProvider, s.printer, checkVersion),
		actions.NewStop(s.ctx, &s.vMProvider, s.printer),
		actions.NewStatus(&s.vMProvider, s.printer),
		actions.NewTerm(s.workingDirectory, &s.ssh.User, &s.ssh.IP, &s.config.UserInfo.Terminal, &s.config.UserInfo.TerminalArguments, &s.vMProvider, s.printer),
		checkVersion,
		unregister.NewUnregister(&s.vMProvider, s.printer),
	)
}
