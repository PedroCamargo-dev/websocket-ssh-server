package utils

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
}

type WSMessage struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Rows    int    `json:"rows,omitempty"`
	Cols    int    `json:"cols,omitempty"`
}

// GetSSHAuthMethods returns a slice of ssh.AuthMethod based on the provided SSHConfig.
// It supports password authentication and private key authentication with or without a passphrase.
//
// Parameters:
//   - config: SSHConfig containing the authentication details.
//
// Returns:
//   - []ssh.AuthMethod: A slice of ssh.AuthMethod to be used for SSH authentication.
//
// If a password is provided in the config, it will be used for password authentication.
// If a private key is provided, it will be parsed and used for public key authentication.
// If both a private key and a password are provided, the private key will be parsed with the passphrase.
func GetSSHAuthMethods(config SSHConfig) []ssh.AuthMethod {
	authMethods := []ssh.AuthMethod{}
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}
	if config.PrivateKey != "" {
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
