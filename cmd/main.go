package main

import (
	"context"
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

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		handlers.HandleWebSocket(ctx, w, r)
	})

	log.Printf("WebSocket running %s/ws", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
