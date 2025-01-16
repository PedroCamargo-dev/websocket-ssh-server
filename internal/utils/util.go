package utils

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	PrivateKey string `json:"privateKey"`
}

type WSMessage struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Rows    int    `json:"rows,omitempty"`
	Cols    int    `json:"cols,omitempty"`
}

func GetSSHAuthMethods(config SSHConfig) []ssh.AuthMethod {
	authMethods := []ssh.AuthMethod{}
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	} else if config.PrivateKey != "" {
		var key ssh.Signer
		var err error

		if config.Password != "" {
			key, err = ssh.ParsePrivateKeyWithPassphrase([]byte(config.PrivateKey), []byte(config.Password))
			if err != nil {
				fmt.Printf("Error parsing private key with passphrase: %v\n", err)
				return authMethods
			}
		} else {
			key, err = ssh.ParsePrivateKey([]byte(config.PrivateKey))
			if err != nil {
				fmt.Printf("Error parsing private key: %v\n", err)
				return authMethods
			}
		}

		authMethods = append(authMethods, ssh.PublicKeys(key))
	}
	return authMethods
}
