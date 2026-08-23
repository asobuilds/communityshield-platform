package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message types
const (
	MessageTypeJoin         = "join"
	MessageTypeLeave        = "leave"
	MessageTypeSignal       = "signal"
	MessageTypeOffer        = "offer"
	MessageTypeAnswer       = "answer"
	MessageTypeIceCandidate = "ice-candidate"
	MessageTypeChat         = "chat"
	MessageTypeLocation     = "location"
	MessageTypeStatus       = "status"
)

// Message represents a WebSocket message
type Message struct {
	Type      string          `json:"type"`
	Room      string          `json:"room"`
	SenderID  string          `json:"senderId"`
	TargetID  string          `json:"targetId,omitempty"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
}

// Client represents a connected WebSocket client
type Client struct {
	ID       string
	Room     string
	Conn     *websocket.Conn
	Send     chan []byte
	UnitID   string
	UserID   string
	Username string
	Role     string
	IsActive bool
}

// Hub maintains the set of active clients
type Hub struct {
	clients    map[string]*Client
	rooms      map[string]map[string]*Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

var hubInstance *Hub
var once sync.Once

// GetHub returns the singleton Hub instance
func GetHub() *Hub {
	once.Do(func() {
		hubInstance = &Hub{
			clients:    make(map[string]*Client),
			rooms:      make(map[string]map[string]*Client),
			broadcast:  make(chan []byte),
			register:   make(chan *Client),
			unregister: make(chan *Client),
		}
		go hubInstance.run()
	})
	return hubInstance
}

// RegisterClient adds a client to the hub
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient removes a client from the hub
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// Broadcast sends a message to all clients in a room
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.ID] = client
			if h.rooms[client.Room] == nil {
				h.rooms[client.Room] = make(map[string]*Client)
			}
			h.rooms[client.Room][client.ID] = client
			h.mutex.Unlock()
			log.Printf("✅ Client %s joined room %s", client.ID, client.Room)
			h.broadcastRoomStatus(client.Room)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				if h.rooms[client.Room] != nil {
					delete(h.rooms[client.Room], client.ID)
					if len(h.rooms[client.Room]) == 0 {
						delete(h.rooms, client.Room)
					}
				}
				close(client.Send)
			}
			h.mutex.Unlock()
			log.Printf("❌ Client %s left room %s", client.ID, client.Room)
			h.broadcastRoomStatus(client.Room)

		case message := <-h.broadcast:
			h.mutex.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client.ID)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (h *Hub) broadcastRoomStatus(room string) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if h.rooms[room] == nil {
		return
	}

	status := struct {
		Type  string   `json:"type"`
		Room  string   `json:"room"`
		Count int      `json:"count"`
		Users []string `json:"users"`
	}{
		Type:  "room-status",
		Room:  room,
		Count: len(h.rooms[room]),
		Users: make([]string, 0),
	}

	for _, client := range h.rooms[room] {
		status.Users = append(status.Users, client.Username)
	}

	msg, _ := json.Marshal(status)
	for _, client := range h.rooms[room] {
		select {
		case client.Send <- msg:
		default:
		}
	}
}

// SendToClient sends a message to a specific client
func (h *Hub) SendToClient(clientID string, message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	if client, ok := h.clients[clientID]; ok {
		select {
		case client.Send <- message:
		default:
		}
	}
}

// SendToRoom sends a message to all clients in a room
func (h *Hub) SendToRoom(room string, message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	if h.rooms[room] == nil {
		return
	}
	for _, client := range h.rooms[room] {
		select {
		case client.Send <- message:
		default:
		}
	}
}
