package handlers

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// Cấu trúc dữ liệu khớp hoàn toàn với Frontend React đang gọi
type SoftwareDisplay struct {
	Tag           string            `json:"tag"`
	Links         map[string]string `json:"links"`
	IsLatest      bool              `json:"is_latest"`
	ReleaseNote   string            `json:"release_note"`
	ReleaseDate   string            `json:"release_date"`
	ChecksumShort string            `json:"checksum_short"`
}

func GetSoftwares(c *gin.Context) {
	// 📂 ĐƯỜNG DẪN THƯ MỤC CHỨA FILE BINARY
	// Thư mục "sof_builds" nằm ngang hàng với file chạy/cmd của bạn
	targetDir := "./sof_builds"

	// Nếu thư mục chưa tồn tại, tự động tạo để tránh lỗi
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		_ = os.MkdirAll(targetDir, os.ModePerm)
	}

	// Đọc toàn bộ danh sách file trong thư mục
	files, err := os.ReadDir(targetDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể đọc thư mục phần mềm: " + err.Error()})
		return
	}

	grouped := make(map[string]*SoftwareDisplay)
	result := make([]*SoftwareDisplay, 0)

	// Quét qua các file để bóc tách thông tin
	// Quy ước tên file bạn bỏ vào folder: SENT_<tag>_<platform><ext>
	// Ví dụ: SENT_v4.2.5_windows.exe, SENT_v4.2.5_linux, SENT_v4.2.5_mac
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		if !strings.HasPrefix(fileName, "SENT_") {
			continue // Bỏ qua file không đúng định dạng hệ thống SENT
		}

		// Lấy thông tin thời gian tạo file làm ngày phát hành
		fileInfo, err := file.Info()
		releaseDate := time.Now().Format("2006-01-02")
		if err == nil {
			releaseDate = fileInfo.ModTime().Format("2006-01-02")
		}

		// Bóc tách chuỗi (Cắt bỏ tiền tố "SENT_")
		cleanName := strings.TrimPrefix(fileName, "SENT_")

		// Xác định Nền tảng (Platform) dựa trên đuôi file
		platform := ""
		tag := ""
		if strings.Contains(cleanName, "_windows.exe") {
			platform = "windows"
			tag = strings.Split(cleanName, "_windows.exe")[0]
		} else if strings.Contains(cleanName, "_linux") {
			platform = "linux"
			tag = strings.Split(cleanName, "_linux")[0]
		} else if strings.Contains(cleanName, "_mac") {
			platform = "mac"
			tag = strings.Split(cleanName, "_mac")[0]
		}

		if platform == "" || tag == "" {
			continue // File không đúng quy ước hệ điều hành
		}

		// Khởi tạo nhóm hiển thị cho Tag này nếu chưa có
		if _, ok := grouped[tag]; !ok {
			vd := &SoftwareDisplay{
				Tag:           tag,
				Links:         make(map[string]string),
				IsLatest:      true, // Mặc định hiển thị là bản mới nhất khi bỏ thủ công vào folder
				ReleaseNote:   "Local build nạp trực tiếp từ thư mục hệ thống.",
				ReleaseDate:   releaseDate,
				ChecksumShort: "LOCAL_MD5",
			}
			grouped[tag] = vd
			result = append(result, vd)
		}

		// Gán link tải trực tiếp trỏ về endpoint static file của Nginx/Server
		grouped[tag].Links[platform] = "https://sent.com.vn/uploads/" + fileName
	}

	c.JSON(http.StatusOK, result)
}

func CreateSoftware(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Tính năng đã bị tắt. Vui lòng copy file trực tiếp vào thư mục sof_builds"})
}

func GetLatestSoftware(c *gin.Context) {
	platform := c.Query("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu thông tin platform"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tag":          "latest",
		"platform":     platform,
		"download_url": "https://sent.com.vn/uploads/SENT_latest_" + platform,
		"is_active":    true,
	})
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
	c.JSON(http.StatusOK, gin.H{"status": "Bỏ qua - Đang chạy chế độ quét folder vật lý"})
}
