package incidents

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
func GetIncidentDetail(c *gin.Context) {
	id := c.Param("id")
	var incident models.Incident

	// [FIX QUAN TRỌNG] Phải Preload cả "Activities" và "Activities.User"
	// Nếu thiếu dòng này, frontend sẽ nhận được mảng activities rỗng -> Không hiện chat/ảnh
	err := database.DB.
		Preload("Agent").           // Lấy thông tin máy trạm
		Preload("Alerts").          // Lấy các cảnh báo gốc
		Preload("Activities.User"). // Lấy thông tin người chat (Avatar, tên)
		Preload("Activities", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at asc") // Sắp xếp tin nhắn từ cũ đến mới
		}).
		First(&incident, id).Error

	if err != nil {
		c.JSON(404, gin.H{"error": "Không tìm thấy sự cố"})
		return
	}

	c.JSON(200, incident)
}

// POST /api/v1/incidents/:id/activity (Thêm ghi chú/Đổi trạng thái)
func AddIncidentActivity(c *gin.Context) {
	id := c.Param("id")

	// 1. [FIX CRITICAL] Lấy User ID an toàn (Chống Crash)
	var userID uint = 0 // Mặc định là 0 (System) nếu không tìm thấy user

	// Thử lấy key "userID" (CamelCase)
	if val, exists := c.Get("userID"); exists && val != nil {
		switch v := val.(type) {
		case uint:
			userID = v
		case float64: // JWT đôi khi trả về float64
			userID = uint(v)
		case int:
			userID = uint(v)
		}
	} else if val, exists := c.Get("user_id"); exists && val != nil {
		// Thử lấy key "user_id" (SnakeCase) dự phòng
		if v, ok := val.(uint); ok {
			userID = v
		}
	}

	// Nếu bắt buộc phải đăng nhập mới được comment, bỏ comment dòng dưới:
	// if userID == 0 {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Vui lòng đăng nhập lại"})
	// 	return
	// }

	// 2. Lấy dữ liệu Text từ Form
	actionType := c.PostForm("action_type")
	content := c.PostForm("content")

	// 3. Xử lý File Upload (Nếu có)
	form, _ := c.MultipartForm()
	var imageURLs models.JSONStringArray // Sử dụng đúng kiểu mảng JSON

	if form != nil {
		files := form.File["files"]
		if len(files) > 0 {
			// Tạo thư mục lưu trữ nếu chưa có
			uploadPath := "uploads/evidence"
			if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
				os.MkdirAll(uploadPath, os.ModePerm)
			}

			for _, file := range files {
				// Tạo tên file an toàn: timestamp_filename
				filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
				dst := filepath.Join(uploadPath, filename)

				if err := c.SaveUploadedFile(file, dst); err == nil {
					// Chuyển đường dẫn thô thành URL web (thay \ bằng / cho chuẩn JSON)
					webPath := "/" + filepath.ToSlash(dst)
					imageURLs = append(imageURLs, webPath)
				}
			}
		}
	}

	// 4. Tìm Sự cố trong DB
	var incident models.Incident
	if err := database.DB.First(&incident, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy sự cố"})
		return
	}

	// 5. Logic chuyển đổi trạng thái
	oldStatus := incident.Status
	newStatus := oldStatus
	if actionType == "RESOLVE" {
		newStatus = "Resolved"
	} else if actionType == "INVESTIGATE" {
		newStatus = "Investigating"
	}

	// 6. Lưu Activity vào DB
	activity := models.IncidentActivity{
		IncidentID: incident.ID,
		UserID:     userID,
		ActionType: actionType,
		Content:    content,
		OldStatus:  oldStatus,
		NewStatus:  newStatus,
		CreatedAt:  time.Now(),
		Images:     imageURLs, // Lưu mảng URL ảnh
	}

	tx := database.DB.Begin()
	if err := tx.Create(&activity).Error; err != nil {
		tx.Rollback()
		fmt.Println("Lỗi lưu Activity:", err)
		c.JSON(500, gin.H{"error": "Lỗi lưu dữ liệu"})
		return
	}

	// Cập nhật trạng thái Incident nếu có thay đổi
	if oldStatus != newStatus {
		if err := tx.Model(&incident).Update("status", newStatus).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "Lỗi cập nhật trạng thái"})
			return
		}
	}
	tx.Commit()

	// Preload User để trả về Frontend hiển thị tên người vừa comment ngay lập tức
	database.DB.Preload("User").First(&activity, activity.ID)

	c.JSON(200, activity)
}

// PUT /api/v1/incidents/:id/playbook
func UpdatePlaybookProgress(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Steps []map[string]interface{} `json:"steps"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Bad Request"})
		return
	}

	// Convert lại thành JSON string để lưu
	jsonBytes, _ := json.Marshal(gin.H{"steps": req.Steps})

	database.DB.Model(&models.Incident{}).Where("id = ?", id).Update("playbook_progress", string(jsonBytes))
	c.JSON(200, gin.H{"message": "Progress saved"})
}

// --- HÀM TỰ ĐỘNG PHÂN LOẠI (Auto-Detect Category) ---
// Hàm này sẽ được dùng cho cả API thêm lẻ và Import Excel sau này
func normalizeCategory(value string, inputCategory string) string {
	// 1. Nếu người dùng đã chọn Category rõ ràng thì tôn trọng
	if inputCategory != "" && inputCategory != "OTHER" {
		return inputCategory
	}

	val := strings.ToUpper(strings.TrimSpace(value))

	// 2. NHẬN DIỆN USB & THIẾT BỊ (Bao gồm cả tên thường gọi)
	// Fix cho trường hợp: "Generic Flash Disk", "Samsung Flash Drive", "VID_..."
	usbKeywords := []string{
		"USB", "VID_", "PID_", "REV_", // Mã kỹ thuật
		"DISK", "DRIVE", "FLASH", "STORAGE", // Tên thương mại
		"KINGSTON", "SANDISK", "SAMSUNG", "TRANSCEND", // Hãng phổ biến
		"DEVICE", "HUB",
	}
	for _, kw := range usbKeywords {
		if strings.Contains(val, kw) {
			return "USB"
		}
	}

	// 3. NHẬN DIỆN PHẦN MỀM
	if strings.HasSuffix(val, ".EXE") || strings.HasSuffix(val, ".MSI") ||
		strings.HasSuffix(val, ".BAT") || strings.HasSuffix(val, ".COM") {
		return "SOFTWARE"
	}

	// 4. NHẬN DIỆN MẠNG (IP hoặc Port)
	// Check nếu là số (Port) hoặc dải IP
	if isNumeric(val) || strings.Contains(val, ".") && !strings.HasSuffix(val, ".EXE") {
		return "NETWORK"
	}

	// Mặc định
	return "OTHER"
}

// Helper check số
func isNumeric(s string) bool {
	match, _ := regexp.MatchString("^[0-9]+$", s)
	return match
}
