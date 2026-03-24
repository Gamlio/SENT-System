// internal/websocket/hub.go
package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	// Lưu trữ kết nối theo HWID: map[HWID]*websocket.Conn
	Clients map[string]*websocket.Conn
	Mu      sync.Mutex
}

var GlobalHub = &Hub{
	Clients: make(map[string]*websocket.Conn),
}

func (h *Hub) Register(hwid string, conn *websocket.Conn) {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	h.Clients[hwid] = conn
}

func (h *Hub) PushCommand(hwid string, message interface{}) {
	h.Mu.Lock()
	conn, exists := h.Clients[hwid]
	h.Mu.Unlock()

	if exists {
		conn.WriteJSON(message)
	}
}
