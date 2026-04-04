package ai

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"sent_backend/internal/database"
	"sent_backend/internal/models"
	service_ai "sent_backend/internal/service/ai"

	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		UserID:    userID,
		Title:     "Cuộc trò chuyện mới",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if database.AIChatSessionCollection == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi kết nối MongoDB"})
		return
	}

	res, err := database.AIChatSessionCollection.InsertOne(context.TODO(), session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi Database"})
		return
	}

	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		session.ID = oid
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

	sessions := make([]models.AIChatSession, 0) // Tránh trả về null
	if database.AIChatSessionCollection != nil {
		opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
		cursor, err := database.AIChatSessionCollection.Find(context.TODO(), bson.M{"user_id": userID}, opts)
		if err == nil {
			cursor.All(context.TODO(), &sessions)
		}
	}
	c.JSON(http.StatusOK, sessions)
}

// POST /api/v1/ai/chat/:session_id
func ChatHandler(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	sessionObjectID, err := primitive.ObjectIDFromHex(sessionIDStr)
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
		SessionID: sessionObjectID,
		Role:      "user",
		Content:   req.Message,
		CreatedAt: time.Now(),
	}
	if database.AIChatLogCollection != nil {
		_, _ = database.AIChatLogCollection.InsertOne(context.TODO(), userLog)
	}

	// Gọi AI
	thought, answer, err := service_ai.ChatWithPolicy(req.Message)
	if err != nil {
		fmt.Println("❌ Lỗi AI:", err) // In lỗi ra server log để debug
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi xử lý AI"})
		return
	}

	// Lưu AI Message
	aiLog := models.AIChatLog{
		SessionID: sessionObjectID,
		UserID:    userID,
		Role:      "ai",
		Content:   answer,
		Thought:   thought,
		CreatedAt: time.Now(),
	}
	if database.AIChatLogCollection != nil {
		_, _ = database.AIChatLogCollection.InsertOne(context.TODO(), aiLog)
	}

	// Update Session
	if database.AIChatSessionCollection != nil {
		_, _ = database.AIChatSessionCollection.UpdateOne(context.TODO(), bson.M{"_id": sessionObjectID}, bson.M{"$set": bson.M{"updated_at": time.Now()}})
	}

	c.JSON(http.StatusOK, gin.H{
		"response": answer,
		"thought":  thought,
	})
}

// --- 1. API ĐỔI TÊN PHIÊN (PUT /api/v1/ai/sessions/:id) ---
func RenameSession(c *gin.Context) {
	sessionIDStr := c.Param("id")
	sessionObjectID, err := primitive.ObjectIDFromHex(sessionIDStr)
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
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu tiêu đề mới"})
		return
	}

	if database.AIChatSessionCollection == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi kết nối MongoDB"})
		return
	}

	res, err := database.AIChatSessionCollection.UpdateOne(
		context.TODO(),
		bson.M{"_id": sessionObjectID, "user_id": userID},
		bson.M{"$set": bson.M{"title": req.Title, "updated_at": time.Now()}},
	)

	if err != nil || res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy phiên hoặc lỗi DB"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đổi tên thành công", "title": req.Title})
}

// --- 2. API XÓA PHIÊN (DELETE /api/v1/ai/sessions/:id) ---
func DeleteSession(c *gin.Context) {
	sessionIDStr := c.Param("id")
	sessionObjectID, err := primitive.ObjectIDFromHex(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID phiên không hợp lệ"})
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if database.AIChatSessionCollection == nil || database.AIChatLogCollection == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi kết nối MongoDB"})
		return
	}

	// Xóa Log
	_, _ = database.AIChatLogCollection.DeleteMany(
		context.TODO(),
		bson.M{"session_id": sessionObjectID, "user_id": userID},
	)

	// Xóa Session
	res, err := database.AIChatSessionCollection.DeleteOne(
		context.TODO(),
		bson.M{"_id": sessionObjectID, "user_id": userID},
	)

	if err != nil || res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy phiên để xóa"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Xóa phiên thành công"})
}

// GET /api/v1/ai/chat/:session_id
func GetChatHistory(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	sessionObjectID, err := primitive.ObjectIDFromHex(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID phiên không hợp lệ"})
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	logs := make([]models.AIChatLog, 0)
	if database.AIChatLogCollection != nil {
		opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
		cursor, err := database.AIChatLogCollection.Find(context.TODO(), bson.M{"session_id": sessionObjectID, "user_id": userID}, opts)
		if err == nil {
			cursor.All(context.TODO(), &logs)
		}
	}

	c.JSON(http.StatusOK, logs)
}
