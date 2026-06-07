package handlers

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"SENT_backend/service/scoring/internal/service" // Import logic tính điểm
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RecalculateAll: Tính toán lại điểm cho toàn bộ máy trạm của tổ chức
func RecalculateAll(c *gin.Context) {
	orgID := c.GetUint("org_id")

	var assetHWIDs []string
	err := database.DB.Model(&models.Asset{}).
		Where("org_id = ?", orgID).
		Pluck("asset_hwid", &assetHWIDs).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách thiết bị"})
		return
	}

	svc := service.NewScoreService(database.DB)
	for _, hwid := range assetHWIDs {
		// Gọi thông qua thực thể struct và truyền đúng chuỗi string hwid
		_ = svc.CalculateRiskScore(c.Request.Context(), hwid)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đã bắt đầu tiến trình tính toán lại điểm số cho toàn bộ thiết bị",
		"count":   len(assetHWIDs),
	})
}

// RecalculateByType: Tính lại điểm khi một loại thiết bị (AssetType) thay đổi trọng số rủi ro
func RecalculateByType(c *gin.Context) {
	orgID := c.GetUint("org_id")
	typeID, _ := strconv.Atoi(c.Param("type_id"))
	typeIDStr := strconv.Itoa(typeID)

	var assetHWIDs []string
	err := database.DB.Model(&models.Asset{}).
		Where("org_id = ? AND asset_type_id = ?", orgID, typeIDStr).
		Pluck("asset_hwid", &assetHWIDs).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lọc danh sách thiết bị"})
		return
	}

	svc := service.NewScoreService(database.DB)
	for _, hwid := range assetHWIDs {
		_ = svc.CalculateRiskScore(c.Request.Context(), hwid)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đã cập nhật lại điểm rủi ro theo loại thiết bị thành công",
		"count":   len(assetHWIDs),
	})
}

// GetScoreHistory: Lấy lịch sử biến thiên rủi ro từ MongoDB
func GetScoreHistory(c *gin.Context) {
	hwid := c.Param("hwid")
	orgID := c.GetUint("org_id")

	// Truy vấn các cảnh báo cũ để xem diễn biến rủi ro
	filter := bson.M{
		"asset_hwid": hwid,
		"org_id":     orgID,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(50) // Lấy 50 bản ghi gần nhất để vẽ biểu đồ

	cursor, err := database.SecurityAlertCollection.Find(c.Request.Context(), filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy lịch sử điểm số"})
		return
	}
	defer cursor.Close(c.Request.Context())

	var history []models.SecurityAlert
	if err = cursor.All(c.Request.Context(), &history); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi xử lý dữ liệu lịch sử"})
		return
	}

	c.JSON(http.StatusOK, history)
}

// HandleRecalculate: Xử lý endpoint POST "/api/v1/scoring/recalculate/:hwid" (Đã đồng bộ tên với main.go)
func HandleRecalculate(c *gin.Context) {
	hwid := c.Param("hwid") // Lấy trực tiếp chuỗi UUID HWID

	if hwid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu thông tin asset_hwid"})
		return
	}

	// Khởi tạo instance ScoreService chuẩn cấu trúc Go
	svc := service.NewScoreService(database.DB)

	// Truyền trực tiếp chuỗi hwid (string) vào hàm CalculateRiskScore
	err := svc.CalculateRiskScore(c.Request.Context(), hwid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi tính toán điểm số rủi ro", "detail": err.Error()})
		return
	}

	// Gọi nội bộ Asset Service để broadcast sự kiện cập nhật (ASSET_UPDATE)
	go func(id string) {
		url := "http://asset-service:8000/api/v1/assets/internal/notify"
		payload := map[string]string{"asset_hwid": id}
		b, _ := json.Marshal(payload)
		_, _ = http.Post(url, "application/json", bytes.NewBuffer(b))
	}(hwid)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Tính toán và cập nhật điểm rủi ro thành công",
		"asset_hwid": hwid,
	})
}
