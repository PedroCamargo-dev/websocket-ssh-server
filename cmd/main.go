package main

import (
	"log"
	"net/http"
	"os"

	"go-websocket-server/internal/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/ws", handlers.HandleWebSocket)
	log.Printf("Servidor WebSocket rodando na porta %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
