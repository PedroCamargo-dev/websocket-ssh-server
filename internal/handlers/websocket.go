package handlers

import (
	"context"
	"log"
	"net/http"

	"go-websocket-server/internal/clients"
	"go-websocket-server/internal/services"
	"go-websocket-server/internal/utils"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket handles the WebSocket connection and communication with the client.
// It upgrades the HTTP connection to a WebSocket connection, reads the initial message from the client,
// starts an SSH session, and handles incoming WebSocket messages.
//
// Parameters:
//   - ctx: The context.Context object for managing the lifecycle of the WebSocket connection.
//   - w: The http.ResponseWriter object for writing the WebSocket response.
//   - r: The *http.Request object representing the WebSocket request.
//
// Returns: None.
//
// Example usage:
//
//	http.HandleFunc("/websocket", HandleWebSocket)
func HandleWebSocket(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	clientID := r.Header.Get("Sec-WebSocket-Key")
	if clientID == "" {
		log.Println("Sec-WebSocket-Key not provided")
		return
	}

	client := &clients.Client{
		Conn:        conn,
		IsConnected: true,
	}
	clients.AddClient(clientID, client)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var msg utils.WSMessage
	if err := conn.ReadJSON(&msg); err != nil {
		log.Printf("Error reading initial message: %v", err)
		return
	}

	session, err := services.StartSSHSession(ctx, msg.Content, conn)
	if err != nil {
		log.Printf("Error starting SSH session: %v", err)
		return
	}
	defer session.Close()

	client.SSHClient = session

	go session.HandleOutput(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Connection terminated by context")
			return
		case <-session.Done:
			log.Println("SSH session ended. Closing WebSocket connection.")
			return
		default:
			if err := conn.ReadJSON(&msg); err != nil {
				log.Printf("Error reading WebSocket message: %v", err)
				return
			}

			switch msg.Type {
			case "input":
				client.Mu.Lock()
				client.CommandBuffer += msg.Content

				if client.CommandBuffer == "exit\r" {
					session.Close()
					client.Mu.Unlock()
					return
				}
				session.SendInput(msg.Content)
				client.Mu.Unlock()
			case "resize":
				session.ResizeTerminal(msg.Rows, msg.Cols)
			default:
				log.Printf("Unknown message type: %s", msg.Type)
			}
		}
	}
}
