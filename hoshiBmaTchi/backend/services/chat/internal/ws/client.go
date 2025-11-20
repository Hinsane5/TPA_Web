package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/chat/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/chat/internal/repositories"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	UserID string
	Repo   *repositories.ChatRepository
}

// ReadPump pumps messages from the websocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Save to DB
		msg := &domain.Message{
			ConversationID: uuid.MustParse(wsMsg.ConversationID),
			SenderID:       uuid.MustParse(c.UserID),
			Content:        wsMsg.Content,
			MediaURL:       wsMsg.MediaURL,
			CreatedAt:      time.Now(),
		}
		
		if err := c.Repo.SaveMessage(context.Background(), msg); err == nil {
			broadcastBytes, _ := json.Marshal(wsMsg)
			c.Hub.broadcast <- broadcastBytes
		}
	}
}

// WritePump pumps messages from the hub to the websocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// In client.go
func ServeWs(hub *Hub, repo *repositories.ChatRepository, w http.ResponseWriter, r *http.Request, userID string) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    client := &Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256), UserID: userID, Repo: repo}
    client.Hub.Register <- client

    go client.WritePump()
    go client.ReadPump()
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    // Allow all origins because Gateway handles CORS
    CheckOrigin: func(r *http.Request) bool { return true },
}