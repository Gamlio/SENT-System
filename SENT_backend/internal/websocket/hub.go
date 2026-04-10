// internal/websocket/hub.go
package websocket

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// SafeConn: Đảm bảo an toàn khi ghi dữ liệu đồng thời
type SafeConn struct {
	Conn *websocket.Conn
	Mu   sync.Mutex
}

// WriteJSON locks the connection and writes a JSON message.
func (s *SafeConn) WriteJSON(v interface{}) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return s.Conn.WriteJSON(v)
}

type Hub struct {
	// Chuyển từ map[string]*websocket.Conn sang map[string]*SafeConn
	Clients map[string]*SafeConn
	Mu      sync.Mutex
}

var GlobalHub = &Hub{
	Clients: make(map[string]*SafeConn),
}
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Cho phép Dashboard kết nối
}

// Khi Register, chúng ta bọc kết nối lại
func (h *Hub) Register(hwid string, conn *websocket.Conn) {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	h.Clients[hwid] = &SafeConn{Conn: conn}
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
	tmpClients := make([]*SafeConn, 0, len(h.Clients))
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

	// CẬP NHẬT: Nếu là Asset kết nối, đánh dấu Online trong Database và báo cho Frontend
	if hwid != "ADMIN_DASHBOARD" {
		// CHỈ cập nhật last_seen, giữ nguyên status (PENDING/ACTIVE)
		database.DB.Model(&models.Asset{}).Where("asset_hwid = ?", hwid).Update("last_seen", time.Now())

		// Phát sự kiện nhưng gửi kèm trạng thái Online tính toán
		GlobalHub.Broadcast(map[string]interface{}{
			"type": "ASSET_HEARTBEAT",
			"data": map[string]string{
				"assetId":    hwid,
				"connection": "online", // Dùng key khác để Frontend hiểu
			},
		})
	}
	// Giữ kết nối mở cho đến khi client tắt hoặc lỗi
	defer func() {
		GlobalHub.Mu.Lock()
		delete(GlobalHub.Clients, hwid)
		GlobalHub.Mu.Unlock()
		conn.Close()

		// CẬP NHẬT: Khi Asset ngắt mạng, đánh dấu Offline và báo cho Frontend
		if hwid != "ADMIN_DASHBOARD" {
			database.DB.Model(&models.Asset{}).Where("asset_hwid = ?", hwid).Update("status", "Offline")
			GlobalHub.Broadcast(map[string]interface{}{
				"type": "ASSET_STATUS_CHANGED",
				"data": map[string]string{
					"assetId": hwid,
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
