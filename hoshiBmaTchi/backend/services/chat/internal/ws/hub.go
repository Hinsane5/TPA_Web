package ws

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type WSMessage struct {
	Type           string `json:"type"`
	ID             string    `json:"id,omitempty"`         
	CreatedAt      time.Time `json:"created_at,omitempty"`
	SenderID       string `json:"sender_id"`
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content"`
	MediaURL       string `json:"media_url,omitempty"`
	MediaType      string `json:"media_type,omitempty"`
}

type Hub struct {
	clients    map[string]*Client
	Register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	redis      *redis.Client
	mu         sync.Mutex
}

func NewHub(rdb *redis.Client) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		redis:      rdb,
	}
}

func (h *Hub) Run() {
	pubsub := h.redis.Subscribe(context.Background(), "chat_broadcast")
	ch := pubsub.Channel()

	go func() {
		for msg := range ch {
			h.sendToLocalClients([]byte(msg.Payload))
		}
	}()

	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.redis.Publish(context.Background(), "chat_broadcast", message)
		}
	}
}

func (h *Hub) sendToLocalClients(msg []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	for _, client := range h.clients {
		select {
		case client.Send <- msg:
		default:
			close(client.Send)
			delete(h.clients, client.UserID)
		}
	}
}

func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}