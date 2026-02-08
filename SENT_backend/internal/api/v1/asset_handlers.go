// internal/api/v1/asset_handlers.go
package v1

import (
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"time"

	"github.com/gin-gonic/gin"
)

func PushDataHandler(c *gin.Context) {
	var req struct {
		EnrollToken string                 `json:"enroll_token"`
		HWID        string                 `json:"hwid"`
		Hostname    string                 `json:"hostname"`
		OSInfo      string                 `json:"os_info"`
		NetworkInfo map[string]interface{} `json:"network_info"`
		USBDevices  []interface{}          `json:"usb_devices"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu sai định dạng"})
		return
	}

	var asset models.Asset
	// 1. Kiểm tra xem Agent này đã tồn tại chưa
	result := database.DB.Where("hwid = ?", req.HWID).First(&asset)

	if result.Error != nil {
		// 2. Nếu là máy mới, tra cứu org_id từ EnrollToken
		var region models.Region
		if err := database.DB.Where("enroll_token = ?", req.EnrollToken).First(&region).Error; err != nil {
			c.JSON(401, gin.H{"error": "Mã Enrollment không hợp lệ"})
			return
		}

		// 3. Tự động gán đúng công ty (Multi-tenant)
		asset = models.Asset{
			HWID:     req.HWID,
			Hostname: req.Hostname,
			OrgID:    region.OrgID, // Tự động định danh công ty
			RegionID: region.ID,
		}
		database.DB.Create(&asset)
	}

	// 4. Cập nhật thông số mới nhất từ máy trạm
	database.DB.Model(&asset).Updates(map[string]interface{}{
		"os_info":      req.OSInfo,
		"network_info": req.NetworkInfo,
		"usb_devices":  req.USBDevices,
		"status":       "online",
		"last_seen":    time.Now(),
	})

	c.JSON(200, gin.H{"status": "Dữ liệu đã được cập nhật cho OrgID: ", "org": asset.OrgID})
}
