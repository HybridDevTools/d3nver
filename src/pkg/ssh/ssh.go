package ssh

import (
	"denver/pkg/util"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	pubkey  = "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEA6NF8iallvQVp22WDkTkyrtvp9eWW6A8YVr+kz4TjGYe7gHzIw+niNltGEFHzD8+v1I2YJ6oXevct1YeS0o9HZyN1Q9qgCgzUFtdOKLv6IedplqoPkcmF0aYet2PkEDo3MlTBckFXPITAMzF8dJSIFo9D8HfdOV0IAdx4O7PtixWKn5y2hMNG0zQPyUecp4pzC6kivAIhyfHilFR61RGL+GPXQ2MWZWFYbAGjyiYJnAmCP3NOTd0jMZEnDkbUvxhMmBYSdETk1rRgm+R4LOzFUGaHqHDLKLX+FIPKcF96hrucXzcWyLbIbEgE98OHlnVYCzRdK8jlqm8tehUc9c9WhQ== ldevuser insecure public key"
	privkey = "-----BEGIN RSA PRIVATE KEY-----\n" +
		"MIIEogIBAAKCAQEA6NF8iallvQVp22WDkTkyrtvp9eWW6A8YVr+kz4TjGYe7gHzI\n" +
		"w+niNltGEFHzD8+v1I2YJ6oXevct1YeS0o9HZyN1Q9qgCgzUFtdOKLv6IedplqoP\n" +
		"kcmF0aYet2PkEDo3MlTBckFXPITAMzF8dJSIFo9D8HfdOV0IAdx4O7PtixWKn5y2\n" +
		"hMNG0zQPyUecp4pzC6kivAIhyfHilFR61RGL+GPXQ2MWZWFYbAGjyiYJnAmCP3NO\n" +
		"Td0jMZEnDkbUvxhMmBYSdETk1rRgm+R4LOzFUGaHqHDLKLX+FIPKcF96hrucXzcW\n" +
		"yLbIbEgE98OHlnVYCzRdK8jlqm8tehUc9c9WhQIBIwKCAQEA4iqWPJXtzZA68mKd\n" +
		"ELs4jJsdyky+ewdZeNds5tjcnHU5zUYE25K+ffJED9qUWICcLZDc81TGWjHyAqD1\n" +
		"Bw7XpgUwFgeUJwUlzQurAv+/ySnxiwuaGJfhFM1CaQHzfXphgVml+fZUvnJUTvzf\n" +
		"TK2Lg6EdbUE9TarUlBf/xPfuEhMSlIE5keb/Zz3/LUlRg8yDqz5w+QWVJ4utnKnK\n" +
		"iqwZN0mwpwU7YSyJhlT4YV1F3n4YjLswM5wJs2oqm0jssQu/BT0tyEXNDYBLEF4A\n" +
		"sClaWuSJ2kjq7KhrrYXzagqhnSei9ODYFShJu8UWVec3Ihb5ZXlzO6vdNQ1J9Xsf\n" +
		"4m+2ywKBgQD6qFxx/Rv9CNN96l/4rb14HKirC2o/orApiHmHDsURs5rUKDx0f9iP\n" +
		"cXN7S1uePXuJRK/5hsubaOCx3Owd2u9gD6Oq0CsMkE4CUSiJcYrMANtx54cGH7Rk\n" +
		"EjFZxK8xAv1ldELEyxrFqkbE4BKd8QOt414qjvTGyAK+OLD3M2QdCQKBgQDtx8pN\n" +
		"CAxR7yhHbIWT1AH66+XWN8bXq7l3RO/ukeaci98JfkbkxURZhtxV/HHuvUhnPLdX\n" +
		"3TwygPBYZFNo4pzVEhzWoTtnEtrFueKxyc3+LjZpuo+mBlQ6ORtfgkr9gBVphXZG\n" +
		"YEzkCD3lVdl8L4cw9BVpKrJCs1c5taGjDgdInQKBgHm/fVvv96bJxc9x1tffXAcj\n" +
		"3OVdUN0UgXNCSaf/3A/phbeBQe9xS+3mpc4r6qvx+iy69mNBeNZ0xOitIjpjBo2+\n" +
		"dBEjSBwLk5q5tJqHmy/jKMJL4n9ROlx93XS+njxgibTvU6Fp9w+NOFD/HvxB3Tcz\n" +
		"6+jJF85D5BNAG3DBMKBjAoGBAOAxZvgsKN+JuENXsST7F89Tck2iTcQIT8g5rwWC\n" +
		"P9Vt74yboe2kDT531w8+egz7nAmRBKNM751U/95P9t88EDacDI/Z2OwnuFQHCPDF\n" +
		"llYOUI+SpLJ6/vURRbHSnnn8a/XG+nzedGH5JGqEJNQsz+xT2axM0/W/CRknmGaJ\n" +
		"kda/AoGANWrLCz708y7VYgAtW2Uf1DPOIYMdvo6fxIB5i9ZfISgcJ/bbCUkFrhoH\n" +
		"+vq/5CIWxCPp0f85R4qxxQ5ihxJ0YDQT9Jpx4TMss4PSavPaBH3RXow5Ohe+bYoQ\n" +
		"NE5OgEXk2wVfZczCZpigBKbKZHNYcelXtTt/nP3rsCuGcM4h53s=\n" +
		"-----END RSA PRIVATE KEY-----"
)

// SSH implements ssh protocol
type SSH struct {
	IP      string
	port    string
	User    string
	keyPath string
}

// NewSSH returns a pointer to SSH
func NewSSH(ip string, workingDirectory string) (*SSH, error) {
	s := &SSH{
		IP:      ip,
		port:    "22",
		User:    "ldevuser",
		keyPath: filepath.Join(workingDirectory, ".ssh"),
	}

	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *SSH) connect() (client *ssh.Client, err error) {
	key, err := ioutil.ReadFile(s.getPrivKey())
	if err != nil {
		return
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return
	}
	config := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
		Timeout:         time.Duration(1) * time.Second,
	}

	client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%s", s.IP, s.port), config)
	return
}

func (s *SSH) init() (err error) {
	keyExist, _ := util.Exists(s.getPrivKey())
	if !keyExist {
		log.Println("Insecure key pair not found, generating a new key pair ...")
		err = s.createKeys()
	}

	return
}

func (s *SSH) createKeys() (err error) {
	if err := os.MkdirAll(s.keyPath, os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(s.getPubKey(), []byte(pubkey), 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(s.getPrivKey(), []byte(privkey), 0600); err != nil {
		return err
	}

	return nil
}

func (s *SSH) getPubKey() string {
	return filepath.Join(s.keyPath, "id_rsa.pub")
}

func (s *SSH) getPrivKey() string {
	return filepath.Join(s.keyPath, "id_rsa")
}
