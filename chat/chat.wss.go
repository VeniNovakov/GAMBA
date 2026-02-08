package chat

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
}

type Hub struct {
	clients    map[uuid.UUID]map[*Client]bool // userID -> clients
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	mu         sync.RWMutex
}

type BroadcastMessage struct {
	UserID  uuid.UUID
	Message []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()
			log.Printf("Client connected: user=%s", client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("Client disconnected: user=%s", client.UserID)

		case msg := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.clients[msg.UserID]; ok {
				for client := range clients {
					select {
					case client.Send <- msg.Message:
					default:
						close(client.Send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) SendToUser(userID uuid.UUID, msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.broadcast <- &BroadcastMessage{
		UserID:  userID,
		Message: data,
	}
}

// Register registers a client
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// IsUserOnline checks if a user has any active connections
func (h *Hub) IsUserOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}

// GetOnlineUsers returns list of online user IDs
func (h *Hub) GetOnlineUsers() []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]uuid.UUID, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

// ReadPump reads messages from the WebSocket connection
func (c *Client) ReadPump(service *Service) {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse incoming message
		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Handle different message types
		switch wsMsg.Type {
		case "send_message":
			c.handleSendMessage(service, wsMsg.Payload)
		case "typing":
			c.handleTyping(service, wsMsg.Payload)
		case "mark_read":
			c.handleMarkRead(service, wsMsg.Payload)
		}
	}
}

// WritePump writes messages to the WebSocket connection
func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.Send
		if !ok {
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

func (c *Client) handleSendMessage(service *Service, payload interface{}) {
	data, _ := json.Marshal(payload)
	var msg WSChatMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	_, err := service.SendMessage(msg.ChatID, c.UserID, &SendMessageRequest{Content: msg.Content})
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (c *Client) handleTyping(service *Service, payload interface{}) {
	data, _ := json.Marshal(payload)
	var event WSTypingEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return
	}

	service.SendTypingIndicator(event.ChatID, c.UserID)
}

func (c *Client) handleMarkRead(service *Service, payload interface{}) {
	data, _ := json.Marshal(payload)
	var event WSReadEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return
	}

	if event.MessageID != uuid.Nil {
		service.MarkAsRead(event.MessageID, c.UserID)
	} else {
		service.MarkChatAsRead(event.ChatID, c.UserID)
	}
}
