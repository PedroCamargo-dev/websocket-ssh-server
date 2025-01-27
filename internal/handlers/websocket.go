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
		appErr := utils.NewAppError("WS_UPGRADE_FAILED", "Failed to upgrade to WebSocket", err)
		appErr.Log()
		http.Error(w, appErr.Message, http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	clientID := r.Header.Get("Sec-WebSocket-Key")
	if clientID == "" {
		appErr := utils.NewAppError("MISSING_CLIENT_ID", "Sec-WebSocket-Key not provided", nil)
		appErr.Log()
		conn.WriteJSON(utils.WSMessage{Type: "error", Content: appErr.Message})
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
		appErr := utils.NewAppError("WS_READ_FAILED", "Failed to read initial WebSocket message", err)
		appErr.Log()
		conn.WriteJSON(utils.WSMessage{Type: "error", Content: appErr.Message})
		return
	}

	session, err := services.StartSSHSession(ctx, msg.Content, conn)
	if err != nil {
		appErr := utils.NewAppError("SSH_SESSION_FAILED", "Failed to start SSH session", err)
		appErr.Log()
		conn.WriteJSON(utils.WSMessage{Type: "error", Content: appErr.Message})
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
				appErr := utils.NewAppError("WS_READ_FAILED", "Failed to read WebSocket message", err)
				appErr.Log()
				conn.WriteJSON(utils.WSMessage{Type: "error", Content: appErr.Message})
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
				err := session.ResizeTerminal(msg.Rows, msg.Cols)
				if err != nil {
					appErr := utils.NewAppError("RESIZE_FAILED", "Failed to resize terminal", err)
					appErr.Log()
					conn.WriteJSON(utils.WSMessage{Type: "error", Content: appErr.Message})
				}
			default:
				appErr := utils.NewAppError("UNKNOWN_MESSAGE_TYPE", "Unknown message type received", nil)
				appErr.Log()
				conn.WriteJSON(utils.WSMessage{Type: "error", Content: appErr.Message})
			}
		}
	}
}
