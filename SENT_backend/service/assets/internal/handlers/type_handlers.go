package handlers

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	assetSvc "SENT_backend/service/assets/internal/service"
	datahelpers "SENT_backend/service/assets/internal/service/data"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAssetTypes(c *gin.Context) {
	orgID := c.GetUint("org_id")
	svc := &assetSvc.AssetTypeService{}
	types, err := svc.GetTypes(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy danh sách"})
		return
	}
	c.JSON(http.StatusOK, types)
}

func GetAssetType(c *gin.Context) {
	orgID := c.GetUint("org_id")
	hwid := c.Param("hwid")

	if hwid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "HWID không được để trống"})
		return
	}

	var asset models.Asset
	if err := database.DB.Where("asset_hwid = ? AND org_id = ?", hwid, orgID).First(&asset).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy tài sản"})
		return
	}

	// If asset has no type assigned, return null
	if asset.AssetTypeID == nil {
		c.JSON(http.StatusOK, gin.H{"asset_type_id": nil})
		return
	}

	// Fetch the asset type
	var assetType models.AssetType
	if err := database.DB.Where("id = ? AND org_id = ?", *asset.AssetTypeID, orgID).First(&assetType).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loại tài sản không tồn tại"})
		return
	}

	c.JSON(http.StatusOK, assetType)
}

func UpdateAssetType(c *gin.Context) {
	orgID := c.GetUint("org_id")
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req models.AssetType
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	svc := &assetSvc.AssetTypeService{}
	updatedType, err := svc.UpdateType(uint(id), orgID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Kích hoạt tính toán lại điểm rủi ro cho tất cả các máy thuộc loại này
	triggerRecalculationByType(uint(id), orgID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật thành công. Hệ thống đang tính toán lại điểm rủi ro cho các thiết bị liên quan.",
		"data":    updatedType,
	})
}

func triggerRecalculationByType(typeID uint, orgID uint) {
	go func() {
		// Lấy danh sách HWID các tài sản thuộc loại này và gọi Scoring Service cho từng máy
		var assetHWIDs []string
		if err := database.DB.Model(&models.Asset{}).
			Where("org_id = ? AND asset_type_id = ?", orgID, typeID).
			Pluck("asset_hwid", &assetHWIDs).Error; err != nil {
			return
		}

		for _, hwid := range assetHWIDs {
			url := "http://scoring-service:8000/api/v1/scoring/recalculate/" + hwid
			// fire-and-forget; scoring handler is unprotected for per-hwid calls
			_, _ = http.Post(url, "application/json", nil)
		}
	}()
}

func CreateAssetType(c *gin.Context) {
	orgID := c.GetUint("org_id")
	var req models.AssetType
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	svc := &assetSvc.AssetTypeService{}
	if err := svc.CreateType(orgID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo loại tài sản mới"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Đã tạo loại tài sản mới thành công"})
}

func AssignAssetTypeToAsset(c *gin.Context) {
	orgID := c.GetUint("org_id")
	hwid := c.Param("hwid")

	var req struct {
		DeviceTypeID *uint `json:"asset_type_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ. 'asset_type_id' phải là một số nguyên."})
		return
	}

	var asset models.Asset
	if err := database.DB.Where("asset_hwid = ? AND org_id = ?", hwid, orgID).First(&asset).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy tài sản."})
		return
	}

	asset.AssetTypeID = req.DeviceTypeID // Assign the new type ID (can be nil to unset)

	if err := database.DB.Save(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lưu thay đổi vào cơ sở dữ liệu."})
		return
	}

	// Use the centralized helper to send a correctly shaped Behavior log (includes nested `asset` object)
	datahelpers.SendBehaviorLog("AssetTypeAssign", "", "Asset type assigned", "Asset type assignment triggered recalculation via Behavior", "P3", asset)

	c.JSON(http.StatusOK, gin.H{"message": "Đã cập nhật phân loại tài sản. Hệ thống đang tính toán lại điểm rủi ro."})
}
