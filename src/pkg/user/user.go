package user

import (
	"denver/pkg/ssh"
	"denver/structs"
	"fmt"
	"os"
)

// User : TODO
type User struct {
	userconf *structs.UserConf
	ssh      *ssh.SSH
}

// NewUser returns a pointer to User
func NewUser(userconf *structs.UserConf, ssh *ssh.SSH) *User {
	return &User{
		userconf: userconf,
		ssh:      ssh,
	}
}

// SetGitUser : TODO
func (u *User) SetGitUser() (err error) {
	if u.userconf.Name == "" {
		return fmt.Errorf("No name configured")
	}

	if _, err = u.ssh.Cmd(fmt.Sprintf("git config --global user.name '%s'", u.userconf.Name)); err != nil {
		return
	}

	if u.userconf.Email == "" {
		return fmt.Errorf("No email configured")
	}

	if _, err = u.ssh.Cmd(fmt.Sprintf("git config --global user.email '%s'", u.userconf.Email)); err != nil {
		return
	}

	return
}

// SetUserKey : TODO
func (u *User) SetUserKey() (err error) {
	if err = u.ssh.Copy(u.userconf.Pubkey, ".ssh/id_rsa.pub", os.FileMode(0644)); err != nil {
		return
	}

	if err = u.ssh.Copy(u.userconf.Privkey, ".ssh/id_rsa", os.FileMode(0600)); err != nil {
		return
	}

	if _, err = u.ssh.Cmd("if [[ $(grep \"`cat .ssh/id_rsa.pub`\" .ssh/authorized_keys &> /dev/null; echo $?) > 0 ]]; then cat .ssh/id_rsa.pub >> .ssh/authorized_keys; else exit 0; fi"); err != nil {
		return
	}

	return
}
