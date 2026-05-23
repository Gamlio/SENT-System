package handlers

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	assetSvc "SENT_backend/service/assets/internal/service"
	"fmt"
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
		url := fmt.Sprintf("http://scoring-service:8000/api/v1/scoring/recalculate/type/%d", typeID)
		resp, err := http.Post(url, "application/json", nil)
		if err != nil {
			log.Printf("Lỗi: Không thể kích hoạt tính toán lại điểm cho loại tài sản %d: %v", typeID, err)
			return
		}
		defer resp.Body.Close()
		log.Printf("Đã kích hoạt tính toán lại điểm cho loại tài sản %d, status: %s", typeID, resp.Status)
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

	// Asynchronously trigger risk score recalculation for this specific asset
	go func(assetHWID string) {
		url := fmt.Sprintf("http://scoring-service:8000/api/v1/scoring/recalculate/%s", assetHWID)
		if _, err := http.Post(url, "application/json", nil); err != nil {
			log.Printf("Lỗi: Không thể kích hoạt tính toán lại điểm cho tài sản %s: %v", assetHWID, err)
		} else {
			log.Printf("Đã kích hoạt tính toán lại điểm cho tài sản %s.", assetHWID)
		}
	}(hwid)

	c.JSON(http.StatusOK, gin.H{"message": "Đã cập nhật phân loại tài sản. Hệ thống đang tính toán lại điểm rủi ro."})
}
