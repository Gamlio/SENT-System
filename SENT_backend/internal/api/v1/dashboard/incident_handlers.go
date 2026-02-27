package dashboard

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	"github.com/gin-gonic/gin"
)

// 1. API: Lấy danh sách Tổng hợp Case (Dùng cho bảng ở Dashboard)
// GET /api/v1/incidents
func GetIncidents(c *gin.Context) {
	db := database.DB
	var incidents []models.Incident

	// Lấy danh sách Incident, sắp xếp mới nhất, kéo theo thông tin Device
	if err := db.Preload("Device").Order("created_at desc").Find(&incidents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách sự cố"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": incidents})
}

// 2. API: Xem Chi tiết 1 Case (Kéo theo Device và toàn bộ Alerts)
// GET /api/v1/incidents/:id
func GetIncidentDetail(c *gin.Context) {
	id := c.Param("id")
	db := database.DB
	var incident models.Incident

	// Preload "Device" và "Alerts" dựa trên struct trong models.go
	if err := db.Preload("Device").Preload("Alerts").First(&incident, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy sự cố"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": incident})
}
