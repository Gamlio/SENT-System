// internal/websocket/hub.go
package websocket

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
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
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Cho phép Dashboard kết nối
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

// Broadcast: Gửi tin nhắn cho TẤT CẢ các client đang kết nối (Dashboard/Admin)
func (h *Hub) Broadcast(message interface{}) {
	h.Mu.Lock()
	tmpClients := make([]*websocket.Conn, 0, len(h.Clients))
	for _, conn := range h.Clients {
		tmpClients = append(tmpClients, conn)
	}
	h.Mu.Unlock() // Mở khóa ngay lập tức

	for _, conn := range tmpClients {
		_ = conn.WriteJSON(message)
	}
}
func WsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	hwid := c.Query("hwid")
	if hwid == "" {
		hwid = "ADMIN_DASHBOARD" // Định danh cho người dùng web
	}

	// Đưa kết nối vào Hub để quản lý
	GlobalHub.Register(hwid, conn)

	// Giữ kết nối mở cho đến khi client tắt hoặc lỗi
	go func() {
		defer func() {
			GlobalHub.Mu.Lock()
			delete(GlobalHub.Clients, hwid)
			GlobalHub.Mu.Unlock()
			conn.Close()
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
