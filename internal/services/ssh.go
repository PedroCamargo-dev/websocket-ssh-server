package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go-websocket-server/internal/clients"
	"go-websocket-server/internal/utils"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type ConfigMessage struct {
	Type    string `json:"type"`
	SSHData struct {
		Host       string `json:"host"`
		Port       string `json:"port"`
		Username   string `json:"username"`
		Password   string `json:"password,omitempty"`
		PrivateKey string `json:"privateKey,omitempty"`
		AuthMethod string `json:"authMethod"`
	} `json:"sshData"`
}

type CommandMessage struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

func NewMessageSender(client *clients.Client) *utils.MessageSender {
	return utils.NewMessageSender(client)
}

// SetupSSHConnection establishes an SSH connection for a given client.
// It retrieves the client using the provided clientID and sends appropriate
// messages based on the connection status.
//
// Parameters:
//   - clientID: A string representing the unique identifier of the client.
//   - config: A ConfigMessage struct containing SSH connection details.
//
// The function performs the following steps:
//  1. Retrieves the client using the clientID.
//  2. Checks if the client is already connected to an SSH server.
//  3. Configures the SSH client using the provided configuration details.
//  4. Establishes the SSH connection using the configured details.
//  5. Updates the client's connection status and SSH client instance.
//  6. Sends appropriate messages to the client regarding the connection status.
//  7. Starts an interactive shell session for the client.
//
// If the client is already connected or if any error occurs during the connection
// process, an error message is sent to the client.
func SetupSSHConnection(clientID string, config ConfigMessage) {
	client := clients.GetClient(clientID)
	messageSender := NewMessageSender(client)

	if client == nil || client.IsConnected {
		messageSender.SendError("Already connected to an SSH server")
		return
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.SSHData.Username,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	if config.SSHData.AuthMethod == "password" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.Password(config.SSHData.Password))
	} else if config.SSHData.AuthMethod == "privateKey" {
		key, err := ssh.ParsePrivateKey([]byte(config.SSHData.PrivateKey))
		if err != nil {
			messageSender.SendError("Invalid private key")
			return
		}
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(key))
	}

	addr := fmt.Sprintf("%s:%s", config.SSHData.Host, config.SSHData.Port)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		messageSender.SendError("Failed to establish SSH connection")
		return
	}

	client.Mu.Lock()
	client.SSHClient = sshClient
	client.IsConnected = true
	client.Mu.Unlock()

	log.Println("Conectado ao servidor SSH")
	messageSender.SendMessage(map[string]string{"type": "connected", "message": "SSH connection established"})
	startShell(clientID)
}

// startShell starts an interactive SSH shell session for the given client ID.
// It retrieves the client from the clients package and initializes a new SSH session.
// If the client or SSH session is nil, it returns immediately.
// The function sets up stdin and stdout pipes for the SSH session and starts a goroutine
// to read from the stdout pipe and send the output to the client's WebSocket connection.
// It also listens for incoming WebSocket messages, unmarshals them into CommandMessage structs,
// and writes the commands to the SSH session's stdin pipe.
//
// Parameters:
//   - clientID: The ID of the client for which to start the SSH shell session.
func startShell(clientID string) {
	client := clients.GetClient(clientID)
	messageSender := NewMessageSender(client)
	if client == nil || client.SSHClient == nil {
		return
	}

	session, err := client.SSHClient.NewSession()
	if err != nil {
		messageSender.SendError("Failed to start shell")
		return
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		messageSender.SendError("Failed to open stdin pipe")
		return
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		messageSender.SendError("Failed to open stdout pipe")
		return
	}

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				log.Println("Shell read error:", err)
				messageSender.SendMessage(map[string]string{"type": "shellClosed", "message": "SSH shell closed"})
				return
			}
			client.Mu.Lock()
			client.Conn.WriteMessage(websocket.TextMessage, buf[:n])
			client.Mu.Unlock()
		}
	}()

	session.RequestPty("xterm-256color", 80, 40, nil)
	session.Shell()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			log.Println("Command read error:", err)
			return
		}
		var cmd CommandMessage
		if err := json.Unmarshal(message, &cmd); err != nil {
			log.Println("Command unmarshal error:", err)
			continue
		}
		stdin.Write([]byte(cmd.Command))
	}
}

// ExecuteCommand sends a command to be executed on the SSH server associated with the given client.
// If the client is nil or not connected, an error message is sent to the client.
//
// Parameters:
//   - client: A pointer to the Client object representing the SSH connection.
//   - command: The command string to be executed on the SSH server.
func ExecuteCommand(client *clients.Client, command string) {
	messageSender := NewMessageSender(client)
	if client == nil || !client.IsConnected {
		messageSender.SendError("No SSH connection available")
		return
	}

	client.SSHClient.Conn.SendRequest(command, false, nil)
}
