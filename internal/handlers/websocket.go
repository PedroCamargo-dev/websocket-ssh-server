package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go-websocket-server/internal/clients"
	"go-websocket-server/internal/services"
	"go-websocket-server/internal/utils"

	"github.com/gorilla/websocket"
)

func NewMessageSender(client *clients.Client) *utils.MessageSender {
	return utils.NewMessageSender(client)
}

// HandleWebSocket handles WebSocket connections by upgrading the HTTP request,
// managing the WebSocket client, and handling incoming messages.
//
// It upgrades the HTTP request to a WebSocket connection, assigns a unique client ID,
// and adds the client to the clients map. It also starts a goroutine to periodically
// send ping messages to keep the connection alive.
//
// Parameters:
//   - w: http.ResponseWriter to send responses to the client.
//   - r: *http.Request containing the client's request.
//
// The function reads messages from the WebSocket connection in a loop and processes
// them using the handleMessage function. If an error occurs during reading or writing
// messages, the connection is cleaned up and closed.
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error connecting to WebSocket:", err)
		return
	}

	clientID := fmt.Sprintf("%d", time.Now().UnixNano())
	client := &clients.Client{Conn: conn}
	clients.AddClient(clientID, client)
	messageSender := NewMessageSender(client)

	log.Println("WebSocket client connected")
	defer clients.CleanupConnection(clientID)

	go func() {
		for {
			time.Sleep(30 * time.Second)
			client.Mu.Lock()
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("Ping error:", err)
				client.Mu.Unlock()
				clients.CleanupConnection(clientID)
				return
			}
			client.Mu.Unlock()
		}
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Erro ao ler mensagem:", err)
			messageSender.SendMessage(map[string]string{"type": "disconnected", "message": "WebSocket connection closed"})
			return
		}

		handleMessage(clientID, message)
	}
}

// handleMessage processes incoming messages for a given client identified by clientID.
// It supports handling configuration messages to set up SSH connections and command messages
// to execute commands on the SSH server.
//
// Parameters:
// - clientID: A string representing the unique identifier of the client.
// - message: A byte slice containing the JSON-encoded message.
//
// The function performs the following steps:
//  1. Retrieves the client associated with the given clientID.
//  2. If the client is not found, the function returns immediately.
//  3. Unmarshals the incoming message into a map to determine its type.
//  4. If the message type is "config", it unmarshals the message into a ConfigMessage
//     and sets up the SSH connection using the provided configuration.
//  5. If the message type is "command", it unmarshals the message into a CommandMessage,
//     appends the command to the client's command buffer, and executes the command if it ends with a newline character.
//  6. If any unmarshaling errors occur, an error message is sent back to the client.
func handleMessage(clientID string, message []byte) {
	client := clients.GetClient(clientID)
	messageSender := NewMessageSender(client)
	if client == nil {
		return
	}

	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		messageSender.SendError("Invalid message format")
		return
	}

	if msg["type"] == "config" {
		var config services.ConfigMessage
		if err := json.Unmarshal(message, &config); err != nil {
			messageSender.SendError("Invalid config format")
			return
		}
		services.SetupSSHConnection(clientID, config)
	} else if msg["type"] == "command" {
		var cmd services.CommandMessage
		if err := json.Unmarshal(message, &cmd); err != nil {
			messageSender.SendError("Invalid command format")
			return
		}
		client.Mu.Lock()
		client.CommandBuffer += cmd.Command
		if cmd.Command == "\n" {
			services.ExecuteCommand(client, client.CommandBuffer)
			client.CommandBuffer = ""
		}
		client.Mu.Unlock()
	}
}
