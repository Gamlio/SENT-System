package handlers

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	softwareService "SENT_backend/service/software/internal/service"
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetSoftwares(c *gin.Context) {
	// Kiểm tra an toàn đề phòng Postgres chưa kết nối xong
	if database.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection is nil"})
		return
	}

	var rawSoftwares []models.SentSoftware
	// Thêm Debug kiểm tra lỗi từ GORM nếu có
	if err := database.DB.Order("created_at desc").Find(&rawSoftwares).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn cơ sở dữ liệu: " + err.Error()})
		return
	}

	type SoftwareDisplay struct {
		Tag           string            `json:"tag"`
		Links         map[string]string `json:"links"`
		IsLatest      bool              `json:"is_latest"`
		ReleaseNote   string            `json:"release_note"`
		ReleaseDate   string            `json:"release_date"`
		ChecksumShort string            `json:"checksum_short"`
	}

	grouped := make(map[string]*SoftwareDisplay)
	result := make([]*SoftwareDisplay, 0)

	for _, v := range rawSoftwares {
		if _, ok := grouped[v.Tag]; !ok {
			// Xử lý cắt chuỗi Checksum an toàn bằng mảng rune (tránh lỗi UTF-8 hoặc rỗng)
			checksumShort := ""
			if v.Checksum != "" {
				runes := []rune(v.Checksum)
				if len(runes) > 8 {
					checksumShort = string(runes[:8])
				} else {
					checksumShort = v.Checksum
				}
			}

			// Kiểm tra định dạng ngày tháng an toàn
			releaseDate := ""
			if !v.CreatedAt.IsZero() {
				releaseDate = v.CreatedAt.Format("2006-01-02")
			}

			vd := &SoftwareDisplay{
				Tag:           v.Tag,
				Links:         make(map[string]string),
				IsLatest:      v.IsLatest,
				ReleaseNote:   v.ReleaseNote,
				ReleaseDate:   releaseDate,
				ChecksumShort: checksumShort,
			}
			grouped[v.Tag] = vd
			result = append(result, vd)
		}
		grouped[v.Tag].Links[v.Platform] = v.DownloadURL
	}
	c.JSON(http.StatusOK, result)
}

func CreateSoftware(c *gin.Context) {
	var input models.SentSoftware
	if err := c.ShouldBindJSON(&input); err != nil {
		println("--- LỖI PARSE JSON WEBHOOK: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	if input.IsLatest {
		database.DB.Model(&models.SentSoftware{}).
			Where("platform = ?", input.Platform).
			Update("is_latest", false)
	}

	if err := database.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lưu cơ sở dữ liệu"})
		return
	}

	c.JSON(http.StatusOK, input)
}

func GetLatestSoftware(c *gin.Context) {
	platform := c.Query("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu thông tin platform"})
		return
	}

	var latest models.SentSoftware
	err := database.DB.Where("platform = ? AND is_latest = ? AND is_active = ?", platform, true, true).
		First(&latest).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy phiên bản mới nhất"})
		return
	}

	c.JSON(http.StatusOK, latest)
}

func GetSoftwareInventory(c *gin.Context) {
	hwid := c.Param("hwid")
	orgID := c.GetUint("org_id")

	var inventory []models.SoftwareItem
	if database.SoftwareCollection != nil {
		cursor, err := database.SoftwareCollection.Find(context.TODO(), bson.M{
			"asset_hwid": hwid,
			"org_id":     orgID,
		})
		if err == nil {
			cursor.All(context.TODO(), &inventory)
		}
	}

	c.JSON(http.StatusOK, inventory)
}

func GitHubWebhookHandler(c *gin.Context) {
	secretToken := os.Getenv("INTERNAL_UPDATE_TOKEN")
	requestToken := c.GetHeader("X-Internal-Token")

	if requestToken == "" || requestToken != secretToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
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

	if err := softwareService.HandleAutoRelease(input.Tag, input.UrlPrefix, input.ReleaseNote); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi cập nhật phiên bản"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Software deployed"})
}
