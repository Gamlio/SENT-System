package version

import (
	"net/http"
	"os"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	versionService "sent_backend/internal/service/version"

	"github.com/gin-gonic/gin"
)

// GetVersions: Lấy danh sách cho Frontend
func GetVersions(c *gin.Context) {
	var versions []models.SentVersion
	database.DB.Order("created_at desc").Find(&versions)
	c.JSON(http.StatusOK, versions)
}

// GitHubWebhookHandler: Nhận lệnh từ GitHub Actions
func GitHubWebhookHandler(c *gin.Context) {
	// Kiểm tra Token bảo mật (đặt trong .env)
	secretToken := os.Getenv("INTERNAL_UPDATE_TOKEN")
	requestToken := c.GetHeader("X-Internal-Token")

	if requestToken == "" || requestToken != secretToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized webhook request"})
		return
	}

	var input struct {
		Tag         string `json:"tag" binding:"required"`
		UrlPrefix   string `json:"url_prefix" binding:"required"`
		ReleaseNote string `json:"release_note"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := versionService.HandleAutoRelease(input.Tag, input.UrlPrefix, input.ReleaseNote); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update version"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Version deployed"})
}
