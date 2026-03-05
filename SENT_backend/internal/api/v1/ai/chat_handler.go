package ai

import (
	"fmt"
	"net/http"
	"strconv"

	"sent_backend/internal/database"
	"sent_backend/internal/models"
	service_ai "sent_backend/internal/service/ai"

	"github.com/gin-gonic/gin"
)

// [SỬA LỖI 401] Hàm này kiểm tra mọi trường hợp có thể của ID
func getUserID(c *gin.Context) (uint, bool) {
	// 1. Thử lấy theo key "userID" (Kiểu CamelCase)
	if val, exists := c.Get("userID"); exists {
		return parseID(val)
	}

	// 2. Thử lấy theo key "user_id" (Kiểu SnakeCase - phòng trường hợp Middleware đặt tên khác)
	if val, exists := c.Get("user_id"); exists {
		return parseID(val)
	}

	// 3. Thử lấy theo key "sub" (Subject trong JWT chuẩn)
	if val, exists := c.Get("sub"); exists {
		return parseID(val)
	}

	// Debug: In ra console xem Middleware đang lưu cái gì
	fmt.Println("⚠️ DEBUG AUTH: Không tìm thấy userID trong Context. Các keys hiện có:", c.Keys)
	return 0, false
}

// Hàm phụ trợ để ép kiểu an toàn
func parseID(val interface{}) (uint, bool) {
	switch v := val.(type) {
	case uint:
		return v, true
	case int:
		return uint(v), true
	case float64: // JWT thường trả về float64
		return uint(v), true
	case string:
		// Trường hợp hãn hữu ID lưu dạng chuỗi
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			return uint(id), true
		}
	}
	return 0, false
}

// POST /api/v1/ai/sessions
func CreateSession(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Lỗi xác thực: Không tìm thấy User ID"})
		return
	}

	session := models.AIChatSession{
		UserID: userID,
		Title:  "Cuộc trò chuyện mới",
	}

	if err := database.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi Database"})
		return
	}
	c.JSON(http.StatusOK, session)
}

// GET /api/v1/ai/sessions
func GetSessions(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var sessions []models.AIChatSession
	database.DB.Where("user_id = ?", userID).Order("updated_at desc").Find(&sessions)
	c.JSON(http.StatusOK, sessions)
}

// POST /api/v1/ai/chat/:session_id
func ChatHandler(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID phiên không hợp lệ"})
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu JSON lỗi"})
		return
	}

	// Lưu User Message
	userLog := models.AIChatLog{
		UserID:    userID,
		SessionID: uint(sessionID),
		Role:      "user",
		Content:   req.Message,
	}
	database.DB.Create(&userLog)

	// Gọi AI
	thought, answer, err := service_ai.ChatWithPolicy(req.Message)
	if err != nil {
		fmt.Println("❌ Lỗi AI:", err) // In lỗi ra server log để debug
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi xử lý AI"})
		return
	}

	// Lưu AI Message
	aiLog := models.AIChatLog{
		SessionID: uint(sessionID),
		UserID:    userID,
		Role:      "ai",
		Content:   answer,
		Thought:   thought,
	}
	database.DB.Create(&aiLog)

	// Update Session
	database.DB.Model(&models.AIChatSession{}).Where("id = ?", sessionID).Update("updated_at", database.DB.NowFunc())

	c.JSON(http.StatusOK, gin.H{
		"response": answer,
		"thought":  thought,
	})
}

// --- 1. API ĐỔI TÊN PHIÊN (PUT /api/v1/ai/sessions/:id) ---
func RenameSession(c *gin.Context) {
	sessionID := c.Param("id")
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu tiêu đề mới"})
		return
	}

	// Chỉ đổi tên nếu session đó thuộc về User này
	result := database.DB.Model(&models.AIChatSession{}).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Update("title", req.Title)

	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy phiên hoặc lỗi DB"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đổi tên thành công", "title": req.Title})
}

// --- 2. API XÓA PHIÊN (DELETE /api/v1/ai/sessions/:id) ---
func DeleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Xóa Session (Gorm dùng Soft Delete nên an toàn)
	// Cascade Delete: Cần cẩn thận, nhưng ở đây ta cứ xóa Session trước
	// Các tin nhắn (AIChatLog) liên quan có thể xóa sau hoặc giữ lại tùy chính sách

	// Cách 1: Xóa cả Session và Log (Sạch sẽ)
	tx := database.DB.Begin()

	// Xóa Log trước
	if err := tx.Where("session_id = ? AND user_id = ?", sessionID, userID).Delete(&models.AIChatLog{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi xóa tin nhắn"})
		return
	}

	// Xóa Session sau
	result := tx.Where("id = ? AND user_id = ?", sessionID, userID).Delete(&models.AIChatSession{})
	if result.Error != nil || result.RowsAffected == 0 {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy phiên để xóa"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Xóa phiên thành công"})
}

// GET /api/v1/ai/chat/:session_id
func GetChatHistory(c *gin.Context) {
	sessionID := c.Param("session_id")
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var logs []models.AIChatLog
	// Lấy tất cả tin nhắn thuộc session và user này, sắp xếp theo thời gian
	database.DB.Where("session_id = ? AND user_id = ?", sessionID, userID).
		Order("created_at asc").
		Find(&logs)

	c.JSON(http.StatusOK, logs)
}
