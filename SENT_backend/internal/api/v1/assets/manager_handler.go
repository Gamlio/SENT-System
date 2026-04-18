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

// GetStats (GATE) - Viết hoa A và gọi service chuẩn hóa
func GetStats(c *gin.Context) {
	orgID := c.GetUint("org_id")
	svc := &assetSvc.AssetDataService{}
	stats, err := svc.GetAssetStats(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy thống kê thiết bị"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetAssets (GATE) - Đổi Getassets -> GetAssets
func GetAssets(c *gin.Context) {
	orgID := c.GetUint("org_id")
	svc := &assetSvc.AssetDataService{}
	c.JSON(http.StatusOK, svc.GetAssetList(orgID))
}

// GetAssetDetail (GATE) - Đổi GetassetDetail -> GetAssetDetail
func GetAssetDetail(c *gin.Context) {
	hwid := c.Param("hwid")
	// [SECURITY] Lấy orgID từ context để đảm bảo đúng phạm vi truy cập
	orgID := c.GetUint("org_id")
	svc := &assetSvc.AssetDataService{}
	asset, err := svc.GetAssetDetail(hwid, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy thiết bị"})
		return
	}
	c.JSON(http.StatusOK, asset)
}

// GetAssetLogs (GATE)
func GetAssetLogs(c *gin.Context) {
	hwid := c.Param("hwid")
	// [SECURITY] Lấy orgID từ context để đảm bảo đúng phạm vi truy cập
	orgID := c.GetUint("org_id")
	svc := &assetSvc.AssetDataService{}
	c.JSON(http.StatusOK, svc.GetAssetLogs(hwid, orgID))
}

func AssignManager(c *gin.Context) {
	hwid := c.Param("hwid")
	var req struct {
		UserID uint `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu UserID hoặc định dạng không hợp lệ"})
		return
	}

	svc := &assetSvc.AssetLifecycleService{}
	// Đảm bảo logic bên trong AssignManager cũng dùng asset_hwid
	if err := svc.AssignManager(hwid, c.GetUint("org_id"), req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã phân công nhân sự"})
}

func RequestDeleteAsset(c *gin.Context) {
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

func RequestBulkDeleteAssets(c *gin.Context) {
	var req struct {
		AssetHWIDs []string `json:"hwids"` // Chuẩn hóa tag JSON cho khớp frontend
		Reason     string   `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return
	}

	username, _ := c.Get("username")
	svc := &assetSvc.AssetLifecycleService{}
	err := svc.CreateBulkDeleteRequest(req.AssetHWIDs, c.GetUint("org_id"), req.Reason, username.(string))
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

	// [SECURITY] Lấy orgID từ context để đảm bảo đúng phạm vi truy cập
	orgID := c.GetUint("org_id")

	svc := &assetSvc.AssetLifecycleService{}
	if err := svc.UpdateDeviceType(hwid, orgID, req.DeviceType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật loại thiết bị thành công"})
}

func TriggerBaseline(c *gin.Context) {
	hwid := c.Param("hwid")
	// [SECURITY] Lấy orgID từ context để đảm bảo đúng phạm vi truy cập
	orgID := c.GetUint("org_id")

	// [FIX-IDOR] Thêm bước xác thực chéo: hwid có thuộc orgID của user không?
	// Đây là chốt chặn quan trọng nhất để chống lại việc user công ty A ra lệnh cho máy công ty B.
	var asset models.Asset
	if err := database.DB.Select("asset_hwid").Where("asset_hwid = ? AND org_id = ?", hwid, orgID).First(&asset).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền truy cập vào máy trạm này hoặc máy trạm không tồn tại."})
		return
	}

	// Kiểm tra xem máy có online không.
	websocket.GlobalHub.Mu.RLock()
	orgClients, orgExists := websocket.GlobalHub.ClientsByOrg[orgID]
	_, clientExists := false, false
	if orgExists {
		_, clientExists = orgClients[hwid]
	}
	websocket.GlobalHub.Mu.RUnlock()

	if !clientExists {
		// Sau khi đã xác thực quyền ở trên, ta có thể chắc chắn lỗi ở đây là do máy offline.
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Máy trạm hiện đang Offline"})
		return
	}

	websocket.GlobalHub.PushCommand(orgID, hwid, gin.H{
		"type": "TRIGGER_BASELINE",
		"data": gin.H{"requester": "Admin"},
	})

	// TRUY VẤN CHUẨN: Đổi hw_id -> asset_hwid
	// [SECURITY] Thêm điều kiện org_id vào câu lệnh update
	database.DB.Model(&models.Asset{}).Where("asset_hwid = ? AND org_id = ?", hwid, orgID).Update("baseline_status", "SCANNING")

	go func(targetAssetHWID string) {
		time.Sleep(5 * time.Second)
		database.DB.Model(&models.Asset{}).Where("asset_hwid = ?", targetAssetHWID).Update("baseline_status", "ESTABLISHED")
	}(hwid)

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi lệnh quét Baseline"})
}
