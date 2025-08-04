package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// TODO: restrict origins in production
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dev
	},
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws", wsHandler)
	log.Println("Starting server on :8080")
	// TODO server with graceful shutdown
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

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Gravity Sling</title>
	</head>
	<body>
		<h1>Welcome to Gravity Sling Multiplayer!</h1>
	<input id="msg" type="text" placeholder="Type a message and press Enter" autofocus />
	<pre id="log"></pre>
	<script>
		const ws = new WebSocket("ws://" + location.host + "/ws");
		const log = document.getElementById("log");
		ws.onopen = () => log.textContent += "WebSocket connected!\n";
		ws.onerror = (err) => log.textContent += "WebSocket error: " + err.message + "\n";
		ws.onclose = () => log.textContent += "WebSocket connection closed\n";
		ws.onmessage = (event) => log.textContent += event.data + "\n";
		document.getElementById("msg").addEventListener("keydown", (event) => {
			if (event.key === "Enter" && event.target.value) {
				console.log("Sending message:", event.target.value);
				ws.send(event.target.value);
				event.target.value = "";
			}
		});
	</script>
	</body>
	</html>
	`)
}
