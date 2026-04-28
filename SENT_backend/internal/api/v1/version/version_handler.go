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
	var rawVersions []models.SentVersion
	database.DB.Order("created_at desc").Find(&rawVersions)

	// Gộp dữ liệu theo Tag để khớp với giao diện VersionCard.jsx
	type VersionDisplay struct {
		Tag           string            `json:"tag"`
		Links         map[string]string `json:"links"`
		IsLatest      bool              `json:"is_latest"`
		ReleaseNote   string            `json:"release_note"`
		ReleaseDate   string            `json:"release_date"`
		ChecksumShort string            `json:"checksum_short"`
	}

	grouped := make(map[string]*VersionDisplay)
	result := make([]*VersionDisplay, 0)

	for _, v := range rawVersions {
		if _, ok := grouped[v.Tag]; !ok {
			checksumShort := v.Checksum
			if len(checksumShort) > 8 {
				checksumShort = checksumShort[:8]
			}

			vd := &VersionDisplay{
				Tag:           v.Tag,
				Links:         make(map[string]string),
				IsLatest:      v.IsLatest,
				ReleaseNote:   v.ReleaseNote,
				ReleaseDate:   v.CreatedAt.Format("2006-01-02"),
				ChecksumShort: checksumShort,
			}
			grouped[v.Tag] = vd
			result = append(result, vd) // Đảm bảo giữ nguyên thứ tự sắp xếp từ DB
		}
		grouped[v.Tag].Links[v.Platform] = v.DownloadURL
	}

	c.JSON(http.StatusOK, result)
}

// GitHubWebhookHandler: Nhận lệnh từ GitHub Actions
func GitHubWebhookHandler(c *gin.Context) {
	// Kiểm tra Token bảo mật (đặt trong .env)
	secretToken := os.Getenv("INTERNAL_UPDATE_TOKEN")
	requestToken := c.GetHeader("X-Internal-Token")

	if requestToken == "" || requestToken != secretToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid Secret Token"})
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
