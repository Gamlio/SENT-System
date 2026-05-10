package handlers

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"SENT_backend/service/scoring/internal/service" // Import logic tính điểm
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
	// Lấy danh sách HWID của tất cả máy thuộc Org này
	err := database.DB.Model(&models.Asset{}).
		Where("org_id = ?", orgID).
		Pluck("asset_hwid", &assetHWIDs).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách thiết bị"})
		return
	}

	// Chạy vòng lặp gọi hàm tính toán có sẵn trong score_service.go
	for _, hwid := range assetHWIDs {
		service.RecalculateRiskScore(hwid)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đã bắt đầu tiến trình tính toán lại điểm cho toàn bộ hệ thống",
		"total":   len(assetHWIDs),
	})
}

// RecalculateByType: Tính lại điểm khi một loại thiết bị (AssetType) thay đổi trọng số rủi ro
func RecalculateByType(c *gin.Context) {
	orgID := c.GetUint("org_id")
	typeID, _ := strconv.Atoi(c.Param("type_id"))

	var assetHWIDs []string
	// Chỉ lấy các máy thuộc loại thiết bị được chỉ định
	err := database.DB.Model(&models.Asset{}).
		Where("asset_type_id = ? AND org_id = ?", typeID, orgID).
		Pluck("asset_hwid", &assetHWIDs).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn danh sách thiết bị theo loại"})
		return
	}

	for _, hwid := range assetHWIDs {
		service.RecalculateRiskScore(hwid)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đã cập nhật lại điểm rủi ro theo loại thiết bị",
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
func HandleRecalculate(c *gin.Context) {
	hwid := c.Param("hwid")
	if hwid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu Asset HWID"})
		return
	}

	// Gọi logic tính toán đã có trong score_service.go
	// Hàm này đã có sẵn cơ chế Debounce để tối ưu hiệu năng
	service.RecalculateRiskScore(hwid)

	c.JSON(http.StatusOK, gin.H{
		"message": "Yêu cầu tính toán lại điểm đã được tiếp nhận",
		"hwid":    hwid,
	})
}
