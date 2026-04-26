package version

import (
	"sent_backend/internal/database"
	"sent_backend/internal/models"
)

// HandleAutoRelease: Tự động cập nhật phiên bản mới từ GitHub Webhook
func HandleAutoRelease(tag string, urlPrefix string, releaseNote string) error {
	db := database.DB
	platforms := []string{"windows", "linux", "mac"}

	// 1. Đánh dấu tất cả bản cũ của các platform này là không còn Latest
	db.Model(&models.SentVersion{}).Where("platform IN ?", platforms).Update("is_latest", false)

	// 2. Tạo bản ghi mới cho từng hệ điều hành
	for _, p := range platforms {
		ext := ""
		if p == "windows" {
			ext = ".exe"
		}

		newVer := models.SentVersion{
			Tag:         tag,
			Platform:    p,
			DownloadURL: urlPrefix + "/SENT_" + p + ext,
			IsLatest:    true,
			IsActive:    true,
			ReleaseNote: releaseNote,
		}
		if err := db.Create(&newVer).Error; err != nil {
			return err
		}
	}
	return nil
}
