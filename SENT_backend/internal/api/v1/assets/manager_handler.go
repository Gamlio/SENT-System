package assets

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	assetSvc "sent_backend/internal/service/assets"
	"sent_backend/internal/websocket"
	"time"

	"github.com/gin-gonic/gin"
)

// GetStats (GATE)
func GetStats(c *gin.Context) {
	svc := &assetSvc.AssetDataService{}
	stats, _ := svc.GetassetStats()
	c.JSON(http.StatusOK, stats)
}

// Getassets (GATE)
func Getassets(c *gin.Context) {
	orgID := c.GetUint("org_id")
	svc := &assetSvc.AssetDataService{}
	c.JSON(http.StatusOK, svc.GetassetList(orgID))
}

// GetassetDetail (GATE)
func GetassetDetail(c *gin.Context) {
	hwid := c.Param("hwid")
	svc := &assetSvc.AssetDataService{}
	asset, err := svc.GetassetDetail(hwid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy thiết bị"})
		return
	}
	c.JSON(http.StatusOK, asset)
}

// GetassetLogs (GATE)
func GetassetLogs(c *gin.Context) {
	hwid := c.Param("hwid")
	svc := &assetSvc.AssetDataService{}
	c.JSON(http.StatusOK, svc.GetassetLogs(hwid))
}

// --- CÁC HÀM THAY ĐỔI TRẠNG THÁI GỌI LIFECYCLE SERVICE ---

func AssignManager(c *gin.Context) {
	hwid := c.Param("hwid")
	var req struct {
		UserID uint `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return
	}

	svc := &assetSvc.AssetLifecycleService{}
	if err := svc.AssignManager(hwid, c.GetUint("org_id"), req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã phân công nhân sự"})
}

func RequestDeleteasset(c *gin.Context) {
	hwid := c.Param("hwid")
	username, _ := c.Get("username")
	orgID := c.GetUint("org_id")

	svc := &assetSvc.AssetLifecycleService{}
	if err := svc.CreateBulkDeleteRequest([]string{hwid}, orgID, "Admin yêu cầu gỡ bỏ", username.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu gỡ bỏ"})
}

// RequestBulkDeleteassets (GATE): Xóa nhiều máy
func RequestBulkDeleteassets(c *gin.Context) {
	var req struct {
		HWIDs  []string `json:"hwids"`
		Reason string   `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return
	}

	username, _ := c.Get("username")
	svc := &assetSvc.AssetLifecycleService{}
	err := svc.CreateBulkDeleteRequest(req.HWIDs, c.GetUint("org_id"), req.Reason, username.(string))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Đã gửi yêu cầu gỡ bỏ hàng loạt"})
}
func UpdateDeviceType(c *gin.Context) {
	hwid := c.Param("hwid")
	var req struct {
		DeviceType string `json:"device_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Gọi Brain xử lý cập nhật và tính lại điểm rủi ro
	svc := &assetSvc.AssetLifecycleService{}
	if err := svc.UpdateDeviceType(hwid, req.DeviceType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật loại thiết bị thành công"})
}

// TriggerBaseline: Ra lệnh cho asset quét sạch hệ thống (Zero Trust)
func TriggerBaseline(c *gin.Context) {
	hwid := c.Param("hwid")

	// 1. Kiểm tra máy có Online không qua Hub
	websocket.GlobalHub.Mu.Lock()
	_, isOnline := websocket.GlobalHub.Clients[hwid]
	websocket.GlobalHub.Mu.Unlock()

	if !isOnline {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Máy trạm hiện đang Offline, không thể nhận lệnh."})
		return
	}

	// 2. Đẩy lệnh xuống asset qua WebSocket
	websocket.GlobalHub.PushCommand(hwid, gin.H{
		"type": "TRIGGER_BASELINE",
		"data": gin.H{"requester": "Admin"},
	})

	// 3. Cập nhật trạng thái Baseline trong DB
	database.DB.Model(&models.Asset{}).Where("hw_id = ?", hwid).Update("baseline_status", "SCANNING")

	// 4. Đặt tiến trình ngầm để tự động reset trạng thái sau khi asset gửi dữ liệu xong
	go func(targetHWID string) {
		time.Sleep(5 * time.Second) // Chờ 5 giây để asset thu thập và gửi dữ liệu về Server
		database.DB.Model(&models.Asset{}).Where("hw_id = ?", targetHWID).Update("baseline_status", "ESTABLISHED")
	}(hwid)

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi lệnh quét Baseline tới máy trạm"})
}
