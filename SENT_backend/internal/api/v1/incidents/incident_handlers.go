package incidents

import (
	"fmt"
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GET /api/v1/incidents
func GetIncidents(c *gin.Context) {
	var incidents []models.Incident

	// 1. In ra màn hình báo là có người gọi API
	fmt.Println("\n--- [DEBUG API] Frontend đang gọi lấy danh sách Incident ---")

	// 2. Thử Query cơ bản nhất (Bỏ Preload tạm thời để xem có phải lỗi quan hệ bảng không)
	result := database.DB.Order("created_at desc").Find(&incidents)

	if result.Error != nil {
		fmt.Printf(">> LỖI DB: %v\n", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy dữ liệu: " + result.Error.Error()})
		return
	}

	// 3. In ra số lượng tìm thấy
	fmt.Printf(">> Tìm thấy: %d sự cố trong Database\n", len(incidents))

	// 4. Nếu có dữ liệu, thử Preload lại Agent để trả về đầy đủ
	if len(incidents) > 0 {
		database.DB.Preload("Agent").Order("created_at desc").Find(&incidents)
	}

	c.JSON(http.StatusOK, gin.H{"data": incidents})
}

// GET /api/v1/incidents/:id
func GetIncidentDetail(c *gin.Context) {
	id := c.Param("id")
	var incident models.Incident

	err := database.DB.Preload("Agent").Preload("Alerts").First(&incident, id).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy sự cố"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": incident})
}
