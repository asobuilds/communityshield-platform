package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"security-solution/models"
	"sync"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client represents a connected WebSocket client
type Client struct {
	Conn     *websocket.Conn
	UserID   uuid.UUID
	UnitID   *uuid.UUID
	CaseID   *uuid.UUID
	Send     chan []byte
	Room     string // "case:{caseID}" or "unit:{unitID}"
}

// Hub manages all clients and rooms
type Hub struct {
	Clients    map[*Client]bool
	Rooms      map[string]map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *Message
	mu         sync.RWMutex
}

// Message represents a chat message
type Message struct {
	ID        string    `json:"id"`
	Room      string    `json:"room"`
	SenderID  string    `json:"senderId"`
	SenderName string   `json:"senderName"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // "chat", "walkie"
}

var hub = Hub{
	Clients:    make(map[*Client]bool),
	Rooms:      make(map[string]map[*Client]bool),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
	Broadcast:  make(chan *Message),
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			if _, ok := h.Rooms[client.Room]; !ok {
				h.Rooms[client.Room] = make(map[*Client]bool)
			}
			h.Rooms[client.Room][client] = true
			h.mu.Unlock()
			log.Printf("✅ Client registered in room: %s", client.Room)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				if _, ok := h.Rooms[client.Room]; ok {
					delete(h.Rooms[client.Room], client)
					if len(h.Rooms[client.Room]) == 0 {
						delete(h.Rooms, client.Room)
					}
				}
				close(client.Send)
			}
			h.mu.Unlock()

		case msg := <-h.Broadcast:
			h.mu.RLock()
			if clients, ok := h.Rooms[msg.Room]; ok {
				for client := range clients {
					select {
					case client.Send <- []byte(msg.ToJSON()):
					default:
						close(client.Send)
						delete(h.Clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (m *Message) ToJSON() string {
	data, _ := json.Marshal(m)
	return string(data)
}

// WebSocketHandler upgrades HTTP to WebSocket
func WebSocketHandler(c *gin.Context) {
	userIDStr := c.Query("userId")
	room := c.Query("room") // "case:{caseID}" or "unit:{unitID}"
	if userIDStr == "" || room == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId and room required"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userId"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		Conn:   conn,
		UserID: userID,
		Room:   room,
		Send:   make(chan []byte, 256),
	}

	hub.Register <- client

	// Read messages from client
	go func() {
		defer func() {
			hub.Unregister <- client
			conn.Close()
		}()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			var message Message
			if err := json.Unmarshal(msg, &message); err != nil {
				continue
			}
			message.ID = uuid.New().String()
			message.Timestamp = time.Now()
			message.SenderID = userID.String()
			message.Room = room

			// Save to DB
			go saveMessageToDB(message)

			hub.Broadcast <- &message
		}
	}()

	// Write messages to client
	go func() {
		defer conn.Close()
		for msg := range client.Send {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				break
			}
		}
	}()
}

func saveMessageToDB(msg Message) {
	var senderID uuid.UUID
	var caseID *uuid.UUID
	var unitID *uuid.UUID

	parsed, err := uuid.Parse(msg.SenderID)
	if err != nil {
		return
	}
	senderID = parsed

	// Determine if it's a case or unit room
	if len(msg.Room) > 5 && msg.Room[:5] == "case:" {
		cid, err := uuid.Parse(msg.Room[5:])
		if err == nil {
			caseID = &cid
		}
	} else if len(msg.Room) > 5 && msg.Room[:5] == "unit:" {
		uid, err := uuid.Parse(msg.Room[5:])
		if err == nil {
			unitID = &uid
		}
	}

	chat := models.ChatMessage{
		SenderID:  senderID,
		CaseID:    caseID,
		UnitID:    unitID,
		Content:   msg.Content,
		Type:      msg.Type,
		Room:      msg.Room,
		Timestamp: msg.Timestamp,
	}
	DB.Create(&chat)
}

// GetChatHistory retrieves messages for a room
func GetChatHistory(c *gin.Context) {
	room := c.Query("room")
	limit := 50
	if room == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room required"})
		return
	}
	var messages []models.ChatMessage
	if err := DB.Where("room = ?", room).Order("timestamp desc").Limit(limit).Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}