package handlers

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	assetSvc "SENT_backend/service/assets/internal/service"
	datahelpers "SENT_backend/service/assets/internal/service/data"
	"log"
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
	total, items := svc.GetAssetList(orgID, page, limit, search, status)

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"limit": limit,
		"items": items,
	})
}

func GetAssetDetail(c *gin.Context) {
	hwid := c.Param("hwid")

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

func BulkAssignManager(c *gin.Context) {
	var req struct {
		HWIDs  []string `json:"hwids" binding:"required"`
		UserID uint     `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu dữ liệu (hwids, user_id) hoặc định dạng không hợp lệ"})
		return
	}

	orgID := c.GetUint("org_id")
	svc := &assetSvc.AssetLifecycleService{}

	for _, hwid := range req.HWIDs {
		if err := svc.AssignManager(hwid, orgID, req.UserID); err != nil {
			log.Printf("Lỗi gán thiết bị %s cho user %d: %v", hwid, req.UserID, err)
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã phân công nhân sự hàng loạt thành công"})
}

func RequestDeleteAsset(c *gin.Context) {
	hwid := c.Param("hwid")
	username, _ := c.Get("username")
	orgID := c.GetUint("org_id")

	// FIX: Cho phép cung cấp lý do trong body của request
	var req struct {
		Reason string `json:"reason"`
	}
	// Binding JSON nhưng không bắt buộc, nếu không có body hoặc reason thì dùng lý do mặc định
	_ = c.ShouldBindJSON(&req)

	reason := "Admin yêu cầu gỡ bỏ"
	if req.Reason != "" {
		reason = req.Reason
	}

	svc := &assetSvc.AssetLifecycleService{}
	if err := svc.CreateDeleteRequest(hwid, orgID, reason, username.(string)); err != nil {
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
	orgID := c.GetUint("org_id")

	if hwid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "HWID không được để trống"})
		return
	}

	if orgID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OrgID không được tìm thấy"})
		return
	}

	var req struct {
		// Sử dụng con trỏ để cho phép gán giá trị null (gỡ bỏ phân loại)
		AssetTypeID *uint `json:"asset_type_id"`
	}

	// Chỉ chấp nhận payload có chứa `asset_type_id`
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ. Payload phải là JSON chứa 'asset_type_id'. Chi tiết: " + err.Error()})
		return
	}

	// Kiểm tra tài sản có tồn tại không
	var asset models.Asset
	if err := database.DB.Where("asset_hwid = ? AND org_id = ?", hwid, orgID).First(&asset).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy tài sản với HWID: " + hwid})
		return
	}

	// KỊCH BẢN 1: Gán một loại tài sản mới (asset_type_id is not null)
	if req.AssetTypeID != nil {
		var assetType models.AssetType
		// Kiểm tra xem ID loại tài sản này có tồn tại và thuộc về tổ chức không
		if err := database.DB.Where("id = ? AND org_id = ?", *req.AssetTypeID, orgID).First(&assetType).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Loại tài sản không tồn tại hoặc không thuộc về tổ chức của bạn."})
			return
		}

		// Cập nhật đồng bộ cả ID liên kết lẫn tên đã được chuẩn hóa
		errUpdate := database.DB.Model(&models.Asset{}).
			Where("asset_hwid = ? AND org_id = ?", hwid, orgID).
			Updates(map[string]interface{}{
				"asset_type_id": *req.AssetTypeID,
				"device_type":   assetType.Name,
			}).Error

		if errUpdate != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi cập nhật phân loại tài sản: " + errUpdate.Error()})
			return
		}

		// Kích hoạt tính toán lại điểm rủi ro
		triggerRiskRecalculation(hwid)

		c.JSON(http.StatusOK, gin.H{"message": "Cập nhật phân loại tài sản thành công.", "device_type": assetType.Name})
		return
	}

	// KỊCH BẢN 2: Gỡ bỏ phân loại (asset_type_id is null)
	errUpdate := database.DB.Model(&models.Asset{}).
		Where("asset_hwid = ? AND org_id = ?", hwid, orgID).
		Updates(map[string]interface{}{
			"asset_type_id": nil,
			"device_type":   nil, // Hoặc gán một giá trị mặc định nếu cần
		}).Error

	if errUpdate != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi gỡ bỏ phân loại tài sản: " + errUpdate.Error()})
		return
	}

	// Kích hoạt tính toán lại điểm rủi ro
	triggerRiskRecalculation(hwid)

	c.JSON(http.StatusOK, gin.H{"message": "Đã gỡ bỏ phân loại tài sản."})
}

// Hàm bổ trợ kích hoạt tính toán lại điểm rủi ro qua API Scoring
func triggerRiskRecalculation(hwid string) {
	go func(id string) {
		// Load asset and call centralized helper to send proper payload
		var asset models.Asset
		if err := database.DB.Where("asset_hwid = ?", id).First(&asset).Error; err != nil {
			return
		}
		datahelpers.SendBehaviorLog("AssetUpdate", "", "Asset metadata changed", "Trigger recalculation via Behavior Service", "P3", asset)
	}(hwid)
}
func UpdateAssetGroup(c *gin.Context) {
	hwid := c.Param("hwid")
	orgID := c.GetUint("org_id")

	if hwid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "HWID không được để trống"})
		return
	}

	if orgID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OrgID không được tìm thấy"})
		return
	}

	var req struct {
		GroupID *uint `json:"group_id"` // Dùng con trỏ để hỗ trợ gán null (Global Policy)
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ. Chi tiết: " + err.Error()})
		return
	}

	// Kiểm tra tài sản có tồn tại không
	var asset models.Asset
	if err := database.DB.Where("asset_hwid = ? AND org_id = ?", hwid, orgID).First(&asset).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy tài sản với HWID: " + hwid})
		return
	}

	// Nếu gán vào nhóm cụ thể, kiểm tra xem nhóm đó có tồn tại thuộc Org không
	if req.GroupID != nil {
		var group models.PolicyGroup
		if err := database.DB.Where("id = ? AND org_id = ?", *req.GroupID, orgID).First(&group).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Nhóm chính sách không tồn tại hoặc không thuộc tổ chức của bạn."})
			return
		}
	}

	errUpdate := database.DB.Model(&models.Asset{}).
		Where("asset_hwid = ? AND org_id = ?", hwid, orgID).
		Update("group_id", req.GroupID).Error

	if errUpdate != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống khi cập nhật nhóm tài sản: " + errUpdate.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Thay đổi nhóm chính sách của tài sản thành công."})
}
