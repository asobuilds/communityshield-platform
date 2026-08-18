package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
	ws "security-solution/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	user, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data"})
		return
	}

	unitID := c.Query("unitId")
	if unitID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unit ID required"})
		return
	}

	var officer models.Officer
	if err := config.DB.Where("user_id = ? AND unit_id = ?", user.ID, unitID).First(&officer).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission for this unit"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}
	defer conn.Close()

	clientID := uuid.New().String()
	hub := ws.GetHub()

	client := &ws.Client{
		ID:       clientID,
		Room:     unitID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UnitID:   unitID,
		UserID:   user.ID.String(),
		Username: user.FirstName + " " + user.LastName,
		Role:     user.Role,
		IsActive: true,
	}

	hub.RegisterClient(client)

	welcome := map[string]interface{}{
		"type":     "welcome",
		"message":  "Connected to walkie-talkie",
		"room":     unitID,
		"clientId": clientID,
	}
	welcomeMsg, _ := json.Marshal(welcome)
	client.Send <- welcomeMsg

	go clientWritePump(client)
	clientReadPump(client, hub)
}

func clientReadPump(client *ws.Client, hub *ws.Hub) {
	defer func() {
		hub.UnregisterClient(client)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(51200)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg ws.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		msg.SenderID = client.ID
		msg.Timestamp = time.Now()

		switch msg.Type {
		case "chat":
			msgData, _ := json.Marshal(msg)
			hub.SendToRoom(client.Room, msgData)

		case "location":
			msgData, _ := json.Marshal(msg)
			hub.SendToRoom(client.Room, msgData)

		case "signal", "offer", "answer", "ice-candidate":
			if msg.TargetID != "" {
				msgData, _ := json.Marshal(msg)
				hub.SendToClient(msg.TargetID, msgData)
			}

		default:
			msgData, _ := json.Marshal(msg)
			hub.SendToRoom(client.Room, msgData)
		}
	}
}

func clientWritePump(client *ws.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}