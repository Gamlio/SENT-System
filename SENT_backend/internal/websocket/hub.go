// internal/websocket/hub.go
package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
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
	// Phân nhóm client theo OrgID để chống rò rỉ tín hiệu (Cross-tenant Signal Leak)
	// map[OrgID] -> map[ClientID] -> *SafeConn
	ClientsByOrg map[uint]map[string]*SafeConn
	Mu           sync.RWMutex // Sử dụng RWMutex để tối ưu cho việc đọc
}

var GlobalHub = &Hub{
	ClientsByOrg: make(map[uint]map[string]*SafeConn),
}
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Cho phép Dashboard kết nối
}

// Register đăng ký một client mới vào Hub, phân loại theo OrgID.
func (h *Hub) Register(orgID uint, clientID string, conn *websocket.Conn) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	if _, ok := h.ClientsByOrg[orgID]; !ok {
		h.ClientsByOrg[orgID] = make(map[string]*SafeConn)
	}
	h.ClientsByOrg[orgID][clientID] = &SafeConn{Conn: conn}
}

// Unregister xóa một client khỏi Hub.
func (h *Hub) Unregister(orgID uint, clientID string) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	if orgClients, ok := h.ClientsByOrg[orgID]; ok {
		delete(orgClients, clientID)
		// Nếu không còn client nào trong Org, xóa luôn entry của Org để giải phóng bộ nhớ
		if len(orgClients) == 0 {
			delete(h.ClientsByOrg, orgID)
		}
	}
}

// PushCommand gửi một lệnh cụ thể đến một client (asset) duy nhất.
// Cần OrgID để tìm kiếm hiệu quả.
func (h *Hub) PushCommand(orgID uint, clientID string, message interface{}) {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	if orgClients, ok := h.ClientsByOrg[orgID]; ok {
		if conn, exists := orgClients[clientID]; exists {
			conn.WriteJSON(message)
		}
	}
}

// BroadcastToOrg gửi tin nhắn cho TẤT CẢ các client thuộc cùng một OrgID.
func (h *Hub) BroadcastToOrg(orgID uint, message interface{}) {
	h.Mu.RLock()
	orgClients, ok := h.ClientsByOrg[orgID]
	if !ok {
		h.Mu.RUnlock()
		return
	}

	// Sao chép danh sách client để tránh giữ lock trong lúc gửi tin
	tmpClients := make([]*SafeConn, 0, len(orgClients))
	for _, conn := range orgClients {
		tmpClients = append(tmpClients, conn)
	}
	h.Mu.RUnlock() // Mở khóa ngay sau khi sao chép xong

	for _, conn := range tmpClients {
		_ = conn.WriteJSON(message)
	}
}

// generateRandomID tạo một ID ngẫu nhiên cho các kết nối từ dashboard.
func generateRandomID(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// WsHandler xử lý các kết nối WebSocket.
// QUAN TRỌŅG: Route này phải được đặt sau middleware `AuthRequired` để có `org_id`.
func WsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	var clientID string
	var orgID uint
	isAsset := false

	// Phân loại kết nối: từ Asset hay từ Dashboard của Admin
	hwid := c.Query("hwid")
	if hwid != "" {
		// Kết nối từ Asset
		isAsset = true
		clientID = hwid
		var asset models.Asset
		if err := database.DB.Select("org_id").Where("asset_hwid = ?", hwid).First(&asset).Error; err != nil {
			conn.Close()
			return
		}
		orgID = asset.OrgID
	} else {
		// Kết nối từ Dashboard (Admin/User)
		// Middleware `AuthRequired` đã xác thực và bơm `org_id` vào context.
		orgIDValue, exists := c.Get("org_id")
		if !exists {
			conn.Close()
			return
		}
		orgIDFromCtx, ok := orgIDValue.(uint)
		if !ok {
			// Điều này không nên xảy ra nếu middleware AuthRequired hoạt động đúng
			conn.Close()
			return
		}
		orgID = orgIDFromCtx
		// Tạo một clientID duy nhất cho mỗi tab trình duyệt của admin
		clientID = fmt.Sprintf("ADMIN_DASHBOARD_%s", generateRandomID(8))
	}

	GlobalHub.Register(orgID, clientID, conn)

	// Nếu là Asset kết nối, cập nhật trạng thái và thông báo cho Org đó
	if isAsset {
		database.DB.Model(&models.Asset{}).Where("asset_hwid = ?", clientID).Update("last_seen", time.Now())
		GlobalHub.BroadcastToOrg(orgID, map[string]interface{}{
			"type": "ASSET_HEARTBEAT",
			"data": map[string]string{
				"assetId":    clientID,
				"connection": "online",
			},
		})
	}

	// Giữ kết nối mở và dọn dẹp khi kết thúc
	defer func() {
		GlobalHub.Unregister(orgID, clientID)
		conn.Close()

		// Nếu là Asset ngắt kết nối, thông báo cho Org đó
		if isAsset {
			// Trạng thái online/offline nên được tính toán động dựa trên `last_seen`.
			GlobalHub.BroadcastToOrg(orgID, map[string]interface{}{
				"type": "ASSET_HEARTBEAT",
				"data": map[string]string{
					"assetId":    clientID,
					"connection": "offline",
				},
			})
		}
	}()

	// Vòng lặp đọc tin nhắn từ client (để phát hiện ngắt kết nối)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
