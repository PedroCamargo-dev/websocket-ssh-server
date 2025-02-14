package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"

	"go-websocket-server/internal/utils"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type SSHSession struct {
	client  *ssh.Client
	session *ssh.Session
	stdin   io.WriteCloser
	stdout  io.Reader
	stderr  io.Reader
	conn    *websocket.Conn
	Done    chan struct{}
	once    sync.Once
}

// StartSSHSession establishes an SSH session with the specified configuration and WebSocket connection.
// It returns a pointer to an SSHSession and an error if any.
// The SSH session is created using the provided SSH configuration in JSON format.
// The SSH authentication methods are determined based on the configuration.
// The SSH client is then connected to the specified host and port.
// A new SSH session is created and configured with a PTY (pseudo-terminal) request.
// The session's standard input, output, and error pipes are opened.
// Finally, the SSHSession struct is initialized with the client, session, pipes, WebSocket connection, and a done channel.
func StartSSHSession(ctx context.Context, configJSON string, conn *websocket.Conn) (*SSHSession, error) {
	var config utils.SSHConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, utils.NewAppError("INVALID_CONFIG", "Invalid configuration format", err)
	}

	authMethods := utils.GetSSHAuthMethods(config)

	clientConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := ssh.Dial("tcp", address, clientConfig)
	if err != nil {
		return nil, utils.NewAppError("SSH_CONNECTION_FAILED", "Failed to connect to SSH server", err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, utils.NewAppError("SSH_SESSION_CREATION_FAILED", "Failed to create SSH session", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return nil, utils.NewAppError("STDIN_PIPE_FAILED", "Failed to open stdin", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, utils.NewAppError("STDOUT_PIPE_FAILED", "Failed to open stdout", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return nil, utils.NewAppError("STDERR_PIPE_FAILED", "Failed to open stderr", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.ECHOCTL:       0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm-256color", 40, 80, modes); err != nil {
		client.Close()
		return nil, utils.NewAppError("PTY_REQUEST_FAILED", "Failed to request PTY", err)
	}

	if err := session.Shell(); err != nil {
		client.Close()
		return nil, utils.NewAppError("SHELL_START_FAILED", "Failed to start shell", err)
	}

	return &SSHSession{
		client:  client,
		session: session,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		conn:    conn,
		Done:    make(chan struct{}),
	}, nil
}

// HandleOutput reads the output from the SSH session's stdout and sends it to the WebSocket connection.
// It continuously reads from the stdout until the context is canceled or the SSH session is done.
// If an error occurs while reading from the stdout, the function will close the SSH session.
// The output is sent to the WebSocket connection as a JSON message with the type "output".
//
// Parameters:
//   - ctx: The context.Context object used to cancel the reading operation.
//
// Returns: None.
func (s *SSHSession) HandleOutput(ctx context.Context) {
	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			log.Println("Closing stdout reading by context")
			return
		case <-s.Done:
			log.Println("Closing stdout reading by done channel")
			return
		default:
			n, err := s.stdout.Read(buf)
			if err != nil {
				if err == io.EOF {
					s.Close()
					return
				}

				appErr := utils.NewAppError("OUTPUT_READ_FAILED", "Failed to read SSH session output", err)
				appErr.Log()
				s.conn.WriteJSON(utils.WSMessage{Type: "error", Content: appErr.Message})
				s.Close()
				return
			}
			s.conn.WriteJSON(utils.WSMessage{Type: "output", Content: string(buf[:n])})
		}
	}
}

// SendInput sends the specified input to the SSH session.
func (s *SSHSession) SendInput(input string) {
	_, err := s.stdin.Write([]byte(input))
	if err != nil {
		log.Printf("Error sending input: %v", err)
	}
}

// ResizeTerminal resizes the terminal window of the SSH session to the specified number of rows and columns.
func (s *SSHSession) ResizeTerminal(rows, cols int) error {
	err := s.session.WindowChange(rows, cols)
	if err != nil {
		log.Printf("Error resizing terminal: %v", err)
		return err
	}
	return nil
}

// Close closes the SSH session and releases any associated resources.
func (s *SSHSession) Close() {
	s.once.Do(func() {
		close(s.Done)
		if s.session != nil {
			s.session.Close()
		}
		if s.client != nil {
			s.client.Close()
		}
		if s.conn != nil {
			s.conn.Close()
		}
	})
}
