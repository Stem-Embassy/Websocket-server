package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

// https://gorilla.github.io/

var clients = make(map[*websocket.Conn]bool)
var mutex sync.Mutex

func main() {
	url := "ws://" + GetLocalIP().String() + ":8080"
	log.Println("- Server starting at: " + url)
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func GetLocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP
}

// SERVER SETUP AND MAIN
func setupRoutes() {
	http.HandleFunc("/", returnHomePage)
	http.HandleFunc("/ws", wsEndpoint)
	http.HandleFunc("/rankings", returnRankings)
}

// HTTTP ENDPOINTS
func returnHomePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
}

func returnRankings(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Rankings")
}

// WEBSOCKET FUNCS
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade Failed", err)
		return
	}

	log.Println("- Client Connected")

	// Use shorter, more reasonable timeouts
	pongWait := 60 * time.Second
	pingInterval := (pongWait * 9) / 10 // Set ping interval to 90% of pong wait time

	// Set pong handler - this is called when we receive a pong from client
	ws.SetPongHandler(func(string) error {
		log.Println("Received pong from client")
		// Reset read deadline whenever we get a pong
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Set ping handler
	ws.SetPingHandler(func(appData string) error {
		log.Println("Received ping from client, sending pong...")
		err := ws.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(10*time.Second))
		if err == nil {
			// Reset read deadline after successful pong
			ws.SetReadDeadline(time.Now().Add(pongWait))
		}
		return err
	})

	// Set initial read deadline
	ws.SetReadDeadline(time.Now().Add(pongWait))

	mutex.Lock()
	clients[ws] = true
	mutex.Unlock()

	// Start a goroutine for sending pings
	go pingClient(ws, pingInterval)
	// Start a goroutine for handling messages
	go handleMessages(ws)
}

func pingClient(ws *websocket.Conn, pingInterval time.Duration) {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		ws.Close() // Ensure connection is closed when this goroutine exits
	}()

	for {
		select {
		case <-ticker.C:
			// Set write deadline for this ping operation
			if err := ws.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				log.Println("SetWriteDeadline error:", err)
				return
			}

			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				log.Println("Ping error:", err)
				// Only close if it's not a normal closure or the connection is already gone
				if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					log.Println("Client unresponsive, closing connection")
				}
				return
			}
			log.Println("Ping sent to client")

			// Reset write deadline after sending ping
			ws.SetWriteDeadline(time.Time{}) // No deadline (blocking mode)
		}
	}
}

// Add this near the towp with other vars
type Message struct {
    Type    int
    Content []byte
    Client  *websocket.Conn
}

var messageQueue = make(chan Message, 100) // Buffer size of 100 messages

func init() {
    // Start the message processor
    go processMessages()
}

func processMessages() {
    for msg := range messageQueue {
        // Process each message synchronously
        log.Println("Processing message:", string(msg.Content))
        
        // Write back to the original sender
        msg.Client.WriteMessage(msg.Type, msg.Content)
        
        // Broadcast to all clients
        mutex.Lock()
        for client := range clients {
            if err := client.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
                log.Println("SetWriteDeadline error:", err)
                client.Close()
                delete(clients, client)
                continue
            }

            err := client.WriteMessage(msg.Type, msg.Content)
            if err != nil {
                log.Println("error broadcasting:", err)
                client.Close()
                delete(clients, client)
            }

            client.SetWriteDeadline(time.Time{})
        }
        mutex.Unlock()
    }
}

func handleMessages(ws *websocket.Conn) {
    defer func() {
        mutex.Lock()
        delete(clients, ws)
        ws.Close()
        log.Println("Closing connection")
        mutex.Unlock()
    }()

    for {
        messageType, p, err := ws.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
                log.Println("Unexpected read error:", err)
            } else {
                log.Println("Connection closed:", err)
            }
            return
        }

        log.Println("Received: ", string(p))
        
        // Queue the message instead of processing it immediately
        messageQueue <- Message{
            Type:    messageType,
            Content: p,
            Client:  ws,
        }
    }
}

// Remove the broadcastMessage function as it's now handled in processMessages
