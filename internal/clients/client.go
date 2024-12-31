package clients

import (
	"sync"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	Conn          *websocket.Conn
	SSHClient     *ssh.Client
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

// CleanupConnection closes the SSH and WebSocket connections for a given client
// and removes the client from the clients map.
//
// Parameters:
//   - clientID: The unique identifier of the client to be cleaned up.
//
// This function locks the clients map to ensure thread safety while performing
// the cleanup operations. If the client exists, it closes the SSH connection
// (if it exists) and the WebSocket connection, then removes the client from
// the clients map.
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
