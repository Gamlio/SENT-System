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
	"sent_backend/internal/service/ai"
	"sent_backend/internal/service/scoring"
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
		Preload("Assignee").        // Lấy thông tin người được phân công (nếu có)
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

	// 1. DÙNG MULTIPART FORM (Cho phép nhận cả text và file ảnh)
	actionType := c.PostForm("action_type")
	content := c.PostForm("content")

	if actionType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu action_type"})
		return
	}

	// 2. LẤY USER ID AN TOÀN TUYỆT ĐỐI (Chống Panic sập Server)
	var uid uint
	if val, exists := c.Get("userID"); exists {
		switch v := val.(type) {
		case float64:
			uid = uint(v)
		case uint:
			uid = v
		case int:
			uid = uint(v)
		}
	} else if val, exists := c.Get("user_id"); exists {
		switch v := val.(type) {
		case float64:
			uid = uint(v)
		case uint:
			uid = v
		case int:
			uid = uint(v)
		}
	}

	// 3. TÌM HỒ SƠ SỰ CỐ
	var incident models.Incident
	if err := database.DB.First(&incident, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy hồ sơ sự cố"})
		return
	}

	oldStatus := incident.Status
	newStatus := oldStatus

	// 4. XỬ LÝ CHUYỂN TRẠNG THÁI & HẠ ĐIỂM RỦI RO
	if actionType == "INVESTIGATE" && oldStatus == "Open" {
		newStatus = "Investigating"
		database.DB.Model(&incident).Updates(map[string]interface{}{
			"status":      newStatus,
			"assignee_id": uid,
		})
	} else if actionType == "RESOLVE" {
		newStatus = "Resolved"
		database.DB.Model(&incident).Update("status", newStatus)
		// Đóng tất cả cảnh báo con
		database.DB.Model(&models.SecurityAlert{}).Where("incident_id = ?", incident.ID).Update("is_resolved", true)
		// Hạ nhiệt điểm rủi ro của máy trạm (Bắt buộc import service scoring)
		scoring.RecalculateRiskScore(incident.AgentHWID)
	}

	// 5. XỬ LÝ LƯU FILE ẢNH VÀO Ổ CỨNG
	var imagePaths []string
	form, err := c.MultipartForm()
	if err == nil {
		files := form.File["images"]
		uploadDir := "uploads/incidents"
		os.MkdirAll(uploadDir, os.ModePerm) // Đảm bảo thư mục tồn tại

		for _, file := range files {
			// Tạo tên file độc nhất tránh trùng lặp
			filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
			filepathStr := filepath.Join(uploadDir, filename)

			if err := c.SaveUploadedFile(file, filepathStr); err == nil {
				// Chuyển đường dẫn thành web path (/uploads/incidents/...)
				webPath := "/" + filepath.ToSlash(filepathStr)
				imagePaths = append(imagePaths, webPath)
			}
		}
	}

	// Convert mảng ảnh thành chuỗi JSON để lưu DB
	imagesJSON, _ := json.Marshal(imagePaths)

	// 6. TẠO LỊCH SỬ TIMELINE
	activity := models.IncidentActivity{
		IncidentID: incident.ID,
		UserID:     &uid,
		ActionType: actionType,
		Content:    content,
		OldStatus:  oldStatus,
		NewStatus:  newStatus,
		Images:     string(imagesJSON), // Lưu mảng ảnh
	}
	database.DB.Create(&activity)

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật thành công", "activity": activity})
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

// PUT /api/v1/incidents/:id/assign
func AssignIncident(c *gin.Context) {
	id := c.Param("id")

	// Lấy UserID an toàn chống Crash
	var userID uint = 0
	if val, exists := c.Get("userID"); exists && val != nil {
		switch v := val.(type) {
		case uint:
			userID = v
		case float64:
			userID = uint(v)
		case int:
			userID = uint(v)
		}
	} else if val, exists := c.Get("user_id"); exists && val != nil {
		if v, ok := val.(uint); ok {
			userID = v
		}
	}

	if userID == 0 {
		c.JSON(401, gin.H{"error": "Không xác định được danh tính. Vui lòng đăng nhập lại!"})
		return
	}

	// Cập nhật người thụ lý và đổi trạng thái sang Investigating
	if err := database.DB.Model(&models.Incident{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"assignee_id": userID,
			"status":      "Investigating",
		}).Error; err != nil {
		c.JSON(500, gin.H{"error": "Lỗi cập nhật"})
		return
	}

	c.JSON(200, gin.H{"message": "Đã nhận xử lý sự cố"})
}
func ExecuteLiveAction(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Command string `json:"command"` // "ISOLATE_NETWORK", "KILL_PROCESS"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu lệnh không hợp lệ"})
		return
	}

	var incident models.Incident
	if err := database.DB.Preload("Agent").First(&incident, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Không tìm thấy sự cố"})
		return
	}

	// Tạo System Log lưu vào DB để truy vết (Ai làm gì, ở đâu)
	logContent := fmt.Sprintf("⚡ LỆNH THỰC THI TỪ XA: [%s]\n> Mục tiêu: %s (%s)\n> Phản hồi: Đã đưa lệnh vào hàng đợi, chờ Agent thực thi...", req.Command, incident.Agent.Hostname, incident.Agent.IPAddress)

	activity := models.IncidentActivity{
		IncidentID: incident.ID,
		UserID:     nil, // nil = Hệ thống/System
		ActionType: "LIVE_RESPONSE",
		Content:    logContent,
		CreatedAt:  time.Now(),
	}
	database.DB.Create(&activity)

	// TODO: Tương lai sẽ gọi MQTT/WebSocket push xuống Agent tại đây.

	c.JSON(200, gin.H{
		"message":  "Đã bắn lệnh xuống thiết bị",
		"activity": activity,
	})
}

// POST /api/v1/incidents/:id/ai-analyze
func AnalyzeIncidentAI(c *gin.Context) {
	id := c.Param("id")

	// Gọi hàm AI trong chat_service.go (nhớ import package chứa hàm đó nếu khác)
	// Giả sử hàm đó nằm trong package `ai`
	analysis, err := ai.AnalyzeIncidentWithAI(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis": analysis,
	})
}
