package handlers

import (
	assetSvc "SENT_backend/service/assets/internal/service"
	"net/http"
	"strconv"

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

func GetAssets(c *gin.Context) {
	orgID := c.GetUint("org_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	status := c.Query("status")

	svc := &assetSvc.AssetDataService{}
	total, items := svc.GetAssetList(orgID, page, limit, search, status) // Đẩy search và status xuống service

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"limit": limit,
		"items": items,
	})
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
	if err := svc.CreateDeleteRequest(hwid, orgID, "Admin yêu cầu gỡ bỏ", username.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu gỡ bỏ máy trạm " + hwid})
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
