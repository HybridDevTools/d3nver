package structs

// InstanceConf : TODO
type InstanceConf struct {
	Name     string
	Provider string
	Vmem     int
	Vcpu     int
	Localip  string
}

// UserConf : TODO
type UserConf struct {
	Name              string
	Email             string
	Pubkey            string
	Privkey           string
	Userdatasize      int
	Terminal          string
	TerminalArguments string
}

// Provider : TODO
type Provider struct {
	Name       string
	Location   string
	Hypervisor string
}

// Config : TODO
type Config struct {
	Channel string
	RBIURL  string
}

// Denver : TODO
type Denver struct {
	Version   string
	Config    *Config
	Instance  *InstanceConf
	UserInfo  *UserConf
	Providers map[string]Provider
}

// NewDenverConfig return pointer to Denver
func NewDenverConfig() *Denver {
	return &Denver{
		Version:   "",
		Config:    &Config{},
		Instance:  &InstanceConf{},
		UserInfo:  &UserConf{},
		Providers: nil,
	}
}
