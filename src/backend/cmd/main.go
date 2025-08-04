package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		// TODO: restrict origins in production
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for dev
		},
	}
	playerNums   = make(map[*websocket.Conn]int)
	playerNumsMu = &sync.Mutex{}
)

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

	playerNum := assignPlayerNum(conn)
	defer removePlayerNum(conn)

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %s\n", err)
			break
		}
		log.Printf("Player %d: %s\n", playerNum, msg)

		if err := conn.WriteMessage(messageType, msg); err != nil {
			log.Printf("Error writing message: %s\n", err)
			break
		}
	}
}

func assignPlayerNum(conn *websocket.Conn) int {
	// Assign the lowest unused player number to the connection
	playerNum := func() int {
		playerNumsMu.Lock()
		defer playerNumsMu.Unlock()
		usedNums := make(map[int]bool)
		for _, num := range playerNums {
			usedNums[num] = true
		}
		playerNum := 1
		for usedNums[playerNum] {
			playerNum++
		}
		playerNums[conn] = playerNum
		return playerNum
	}()
	log.Printf("Assigned player number %d to connection %s\n", playerNum, conn.RemoteAddr())
	if err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("You are player %d", playerNum))); err != nil {
		log.Printf("Error sending player number to connection %s: %s\n", conn.RemoteAddr(), err)
	}
	return playerNum
}

func removePlayerNum(conn *websocket.Conn) {
	playerNumsMu.Lock()
	defer playerNumsMu.Unlock()
	delete(playerNums, conn)
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
