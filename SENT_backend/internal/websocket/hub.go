// internal/websocket/hub.go
package websocket

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
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

	// CẬP NHẬT: Nếu là Agent kết nối, đánh dấu Online trong Database và báo cho Frontend
	if hwid != "ADMIN_DASHBOARD" {
		// Lưu vào Database
		database.DB.Model(&models.Agent{}).Where("hw_id = ?", hwid).Update("status", "Online")

		// Phát sự kiện Real-time để giao diện web đổi màu xanh lập tức
		GlobalHub.Broadcast(map[string]interface{}{
			"type": "AGENT_STATUS_CHANGED",
			"data": map[string]string{
				"agentId": hwid,
				"status":  "Online",
			},
		})
	}

	// Giữ kết nối mở cho đến khi client tắt hoặc lỗi
	defer func() {
		GlobalHub.Mu.Lock()
		delete(GlobalHub.Clients, hwid)
		GlobalHub.Mu.Unlock()
		conn.Close()

		// CẬP NHẬT: Khi Agent ngắt mạng, đánh dấu Offline và báo cho Frontend
		if hwid != "ADMIN_DASHBOARD" {
			database.DB.Model(&models.Agent{}).Where("hw_id = ?", hwid).Update("status", "Offline")
			GlobalHub.Broadcast(map[string]interface{}{
				"type": "AGENT_STATUS_CHANGED",
				"data": map[string]string{
					"agentId": hwid,
					"status":  "Offline",
				},
			})
		}
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
