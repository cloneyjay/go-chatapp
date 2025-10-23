package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Message represents a chat message exchanged between clients
type Message struct {
	Type     string    `json:"type"`     // Type of message (e.g., "chat", "join", "leave")
	Username string    `json:"username"` // Username of the sender
	Text     string    `json:"text"`     // Message content
	Time     time.Time `json:"time"`     // Timestamp of the message
}

// Client represents a connected WebSocket client
type Client struct {
	hub      *Hub            // Reference to the hub
	conn     *websocket.Conn // WebSocket connection
	send     chan []byte     // Buffered channel for outbound messages
	username string          // Username of the client
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	clients    map[*Client]bool // Registered clients
	broadcast  chan []byte      // Inbound messages from clients
	register   chan *Client     // Register requests from clients
	unregister chan *Client     // Unregister requests from clients
}

// newHub creates a new Hub instance
func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// run starts the hub's main loop to handle client registration, unregistration, and message broadcasting
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			// Register new client
			h.clients[client] = true
			log.Printf("Client registered: %s (Total clients: %d)", client.username, len(h.clients))

		case client := <-h.unregister:
			// Unregister client if it exists
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client unregistered: %s (Total clients: %d)", client.username, len(h.clients))

				// Broadcast leave message to all remaining clients
				leaveMsg := Message{
					Type:     "leave",
					Username: client.username,
					Text:     client.username + " has left the chat",
					Time:     time.Now(),
				}
				if msgBytes, err := json.Marshal(leaveMsg); err == nil {
					for c := range h.clients {
						select {
						case c.send <- msgBytes:
						default:
							close(c.send)
							delete(h.clients, c)
						}
					}
				}
			}

		case message := <-h.broadcast:
			// Broadcast message to all connected clients
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// If we can't send to a client, close and remove it
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow connections from any origin (for development)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse the incoming message
		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Set the username and timestamp
		msg.Username = c.username
		msg.Time = time.Now()

		// Marshal back to JSON and broadcast
		if msgBytes, err := json.Marshal(msg); err == nil {
			c.hub.broadcast <- msgBytes
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles WebSocket requests from clients
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Get username from query parameter
	username := r.URL.Query().Get("username")
	if username == "" {
		username = "Anonymous"
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create new client
	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		username: username,
	}

	// Register the client with the hub
	client.hub.register <- client

	// Broadcast join message to all clients
	joinMsg := Message{
		Type:     "join",
		Username: username,
		Text:     username + " has joined the chat",
		Time:     time.Now(),
	}
	if msgBytes, err := json.Marshal(joinMsg); err == nil {
		hub.broadcast <- msgBytes
	}

	// Start goroutines for reading and writing
	// writePump is started first to ensure the client is ready to receive messages
	go client.writePump()
	go client.readPump()
}

func main() {
	// Create and start the hub
	hub := newHub()
	go hub.run()

	// Setup HTTP routes
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	// Serve static files from the public directory
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	// Start the server
	addr := ":8080"
	log.Printf("Chat server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
