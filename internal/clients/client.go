package clients

import (
	"go-websocket-server/internal/services"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn          *websocket.Conn
	SSHClient     *services.SSHSession
	IsConnected   bool
	CommandBuffer string
	Mu            sync.Mutex
}

var (
	clients = make(map[string]*Client)
	mu      sync.Mutex
)

func GetClient(clientID string) *Client {
	mu.Lock()
	defer mu.Unlock()
	return clients[clientID]
}

func AddClient(clientID string, client *Client) {
	mu.Lock()
	clients[clientID] = client
	mu.Unlock()
}

// CleanupConnection cleans up the connection for a given client.
// It closes the SSH client and the connection associated with the client.
// If the client does not exist, it does nothing.
//
// Parameters:
// - clientID: The ID of the client to clean up the connection for.
//
// Example usage:
//
//	CleanupConnection("client-123")
//
// FILEPATH: /home/pedrocamargo/projects/ssh-web-based/websocket-ssh-server/internal/clients/client.go
func CleanupConnection(clientID string) {
	mu.Lock()
	client, exists := clients[clientID]
	if exists {
		if client.SSHClient != nil {
			client.SSHClient.Close()
		}
		client.Conn.Close()
		delete(clients, clientID)
	}
	mu.Unlock()
}
