package handlers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	service_ai "SENT_backend/service/ai/internal/service"

	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getUserID(c *gin.Context) (uint, bool) {
	if val, exists := c.Get("userID"); exists {
		return parseID(val)
	}

	if val, exists := c.Get("user_id"); exists {
		return parseID(val)
	}

	if val, exists := c.Get("sub"); exists {
		return parseID(val)
	}

	fmt.Println("⚠️ DEBUG AUTH: Không tìm thấy userID trong Context. Các keys hiện có:", c.Keys)
	return 0, false
}

func parseID(val interface{}) (uint, bool) {
	switch v := val.(type) {
	case uint:
		return v, true
	case int:
		return uint(v), true
	case float64:
		return uint(v), true
	case string:

		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			return uint(id), true
		}
	}
	return 0, false
}

func extractAIResponseText(rawChunk map[string]interface{}) string {
	if response, ok := rawChunk["response"].(string); ok && response != "" {
		return response
	}
	if candidates, ok := rawChunk["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if candidateMap, ok := candidates[0].(map[string]interface{}); ok {
			if content, ok := candidateMap["content"].(string); ok {
				return content
			}
		}
	}
	if text, ok := rawChunk["text"].(string); ok {
		return text
	}
	if output, ok := rawChunk["output"].(string); ok {
		return output
	}
	return ""
}

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

func GetSessions(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	sessions := make([]models.AIChatSession, 0)
	if database.AIChatSessionCollection != nil {
		opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
		cursor, err := database.AIChatSessionCollection.Find(context.TODO(), bson.M{"user_id": userID}, opts)
		if err == nil {
			cursor.All(context.TODO(), &sessions)
		}
	}
	c.JSON(http.StatusOK, sessions)
}
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

	orgID := c.GetUint("org_id")

	var req struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu JSON lỗi"})
		return
	}

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

	stream, err := service_ai.StreamChatWithRAG(c.Request.Context(), req.Message, orgID)
	if err != nil {
		fmt.Println("❌ Lỗi AI:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi xử lý AI"})
		return
	}
	defer stream.Close()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	var fullAIResponse string
	reader := bufio.NewReader(stream)

	c.Stream(func(w io.Writer) bool {
		line, err := reader.ReadBytes('\n')
		if err != nil && len(line) == 0 {
			return false
		}
		cleanedLine := bytes.TrimSpace(line)
		if len(cleanedLine) == 0 {
			if err != nil {
				return false
			}
			return true
		}
		var rawChunk map[string]interface{}
		if err := json.Unmarshal(cleanedLine, &rawChunk); err != nil {
			fmt.Printf("⚠️ Lỗi giải mã dòng JSON từ AI: %v | Data: %s\n", err, string(cleanedLine))
			if err != nil {
				return false
			}
			return true
		}

		responseText := extractAIResponseText(rawChunk)
		isDone, _ := rawChunk["done"].(bool)

		fullAIResponse += responseText

		c.SSEvent("message", gin.H{
			"text": responseText,
			"done": isDone,
		})

		if isDone {
			return false
		}

		if err != nil {
			return false
		}

		return true
	})
	thought, answer := service_ai.ParseAIResponse(fullAIResponse)

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

	if database.AIChatSessionCollection != nil {
		_, _ = database.AIChatSessionCollection.UpdateOne(
			context.TODO(),
			bson.M{"_id": sessionObjectID},
			bson.M{"$set": bson.M{"updated_at": time.Now()}},
		)
	}
}

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

	_, _ = database.AIChatLogCollection.DeleteMany(
		context.TODO(),
		bson.M{"session_id": sessionObjectID, "user_id": userID},
	)

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

func UploadPlaybook(c *gin.Context) {
	file, err := c.FormFile("playbook")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Không tìm thấy file 'playbook' trong request"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể mở file đã upload"})
		return
	}
	defer f.Close()

	mdContent, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể đọc nội dung file"})
		return
	}

	err = service_ai.IngestPlaybookMarkdownToVectorDB(string(mdContent))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Lỗi nạp playbook: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nạp playbook vào vector DB thành công."})
}
