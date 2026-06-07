package handlers

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	assetSvc "SENT_backend/service/assets/internal/service"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

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
	// Giả định: svc.UpdateType đã được sửa để trả về (updatedObject, error)
	updatedType, err := svc.UpdateType(uint(id), orgID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Kích hoạt tính toán lại điểm rủi ro cho tất cả các máy thuộc loại này
	triggerRecalculationByType(uint(id))

	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật thành công. Hệ thống đang tính toán lại điểm rủi ro cho các thiết bị liên quan.",
		"data":    updatedType,
	})
}

func triggerRecalculationByType(typeID uint) {
	go func() {
		time.Sleep(1 * time.Second) // Đợi 1 chút để DB commit xong
		payload := map[string]interface{}{
			"org_id":   0,
			"type_id":  typeID,
			"category": "AssetTypeChange",
			"title":    "Asset type updated",
			"desc":     "Trigger centralized scoring via Behavior Service",
		}
		b, _ := json.Marshal(payload)
		url := "http://behavior-service:8000/api/v1/behaviors/log"
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(b))
		if err != nil {
			log.Printf("Lỗi: Không thể thông báo Behavior Service cho loại tài sản %d: %v", typeID, err)
			return
		}
		defer resp.Body.Close()
		log.Printf("Đã thông báo Behavior Service cho loại tài sản %d, status: %s", typeID, resp.Status)
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

	// Asynchronously notify Behavior Service; Behavior will trigger centralized scoring
	go func(assetHWID string) {
		payload := map[string]interface{}{
			"org_id":        orgID,
			"asset_hwid":    assetHWID,
			"category":      "AssetTypeAssign",
			"value":         "",
			"title":         "Asset type assigned",
			"desc":          "Asset type assignment triggered recalculation via Behavior",
			"base_priority": "P3",
		}
		b, _ := json.Marshal(payload)
		url := "http://behavior-service:8000/api/v1/behaviors/log"
		if resp, err := http.Post(url, "application/json", bytes.NewBuffer(b)); err != nil {
			log.Printf("Lỗi: Không thể thông báo Behavior Service cho tài sản %s: %v", assetHWID, err)
		} else {
			resp.Body.Close()
			log.Printf("Đã thông báo Behavior Service cho tài sản %s.", assetHWID)
		}
	}(hwid)

	c.JSON(http.StatusOK, gin.H{"message": "Đã cập nhật phân loại tài sản. Hệ thống đang tính toán lại điểm rủi ro."})
}
