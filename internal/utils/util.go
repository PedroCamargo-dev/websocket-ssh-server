package utils

import (
	"sync"

	"go-websocket-server/internal/clients"
)

type MessageSender struct {
	client *clients.Client
	mu     sync.Mutex
}

func NewMessageSender(client *clients.Client) *MessageSender {
	return &MessageSender{
		client: client,
	}
}

func (ms *MessageSender) SendMessage(message map[string]string) {
	if ms.client == nil {
		return
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.client.Conn.WriteJSON(message)
}

func (ms *MessageSender) SendError(message string) {
	if ms.client == nil {
		return
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.client.Conn.WriteJSON(map[string]string{"type": "error", "message": message})
}
