package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dev
	},
}

func main() {
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/ws", wsHandler)
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %s\n", err)
		return
	}
	defer conn.Close()

	log.Println("WebSocket connection established")
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %s\n", err)
			break
		}
		log.Printf("Received message: %s\n", msg)

		if err := conn.WriteMessage(messageType, msg); err != nil {
			log.Printf("Error writing message: %s\n", err)
			break
		}
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World! Let's make a %s game", r.URL.Path[1:])
}
