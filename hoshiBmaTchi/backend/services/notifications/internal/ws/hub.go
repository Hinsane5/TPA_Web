package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients map[uint][]*websocket.Conn
	lock    sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[uint][]*websocket.Conn),
	}
}

func (h *Hub) Register(userID uint, conn *websocket.Conn) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.clients[userID] = append(h.clients[userID], conn)
	log.Printf("User %d connected. Total clients: %d", userID, len(h.clients))
}

func (h *Hub) Unregister(userID uint, conn *websocket.Conn) {
	h.lock.Lock()
	defer h.lock.Unlock()
	
	conns := h.clients[userID]
	for i, c := range conns {
		if c == conn {
			h.clients[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
	if len(h.clients[userID]) == 0 {
		delete(h.clients, userID)
	}
}

func (h *Hub) SendNotification(userID uint, payload interface{}) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	conns, ok := h.clients[userID]
	if !ok {
		return 
	}

	data, _ := json.Marshal(payload)

	for _, conn := range conns {
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("WS Write Error: %v", err)
			conn.Close()
		}
	}
}