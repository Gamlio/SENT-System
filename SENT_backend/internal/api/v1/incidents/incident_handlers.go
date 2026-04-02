package incidents

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/ai"
	incidentService "sent_backend/internal/service/incidents"
	"sent_backend/internal/service/scoring"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// GET /api/v1/incidents
func GetIncidents(c *gin.Context) {
	orgID := c.GetUint("org_id")
	fmt.Println("\n--- [DEBUG API] Frontend đang gọi lấy danh sách Incident ---")

	svc := incidentService.NewIncidentService(database.DB)
	incidents, err := svc.GetAllIncidents(orgID)
	if err != nil {
		fmt.Printf(">> LỖI DB: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy dữ liệu: " + err.Error()})
		return
	}

	fmt.Printf(">> Tìm thấy: %d sự cố trong Database\n", len(incidents))
	c.JSON(http.StatusOK, gin.H{"data": incidents})
}
func GetIncidentDetail(c *gin.Context) {
	id := c.Param("id")
	incidentID := parseUint(id)
	if incidentID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	svc := incidentService.NewIncidentService(database.DB)
	incident, err := svc.GetIncidentByID(incidentID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Không tìm thấy sự cố"})
		return
	}

	c.JSON(200, incident)
}

// POST /api/v1/incidents/:id/activity (Thêm ghi chú/Đổi trạng thái)
func AddIncidentActivity(c *gin.Context) {
	id := c.Param("id")

	// [QUAN TRỌNG NHẤT]: XÓA BỎ ShouldBindJSON.
	// Ép Gin đọc Multipart Form (Tối đa 32MB) để nhận được cả Text lẫn Ảnh
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		fmt.Println("Lỗi đọc form:", err)
	}

	// Đọc Text từ FormData
	actionType := c.PostForm("action_type")
	content := c.PostForm("content")

	// Nếu vẫn trống nghĩa là Frontend gửi sai
	if actionType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Backend không nhận được action_type."})
		return
	}

	// Lấy UserID an toàn chống Crash
	var uid uint
	if val, exists := c.Get("userID"); exists && val != nil {
		switch v := val.(type) {
		case uint:
			uid = v
		case float64:
			uid = uint(v)
		case int:
			uid = uint(v)
		}
	} else if val, exists := c.Get("user_id"); exists && val != nil {
		if v, ok := val.(uint); ok {
			uid = v
		}
	}

	// Tìm Hồ sơ Sự cố
	var incident models.Incident
	if err := database.DB.First(&incident, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy hồ sơ sự cố"})
		return
	}

	oldStatus := incident.Status
	newStatus := oldStatus

	// --- LOGIC XỬ LÝ CHUYỂN TRẠNG THÁI ---
	if actionType == "INVESTIGATE" && oldStatus == "Open" {
		newStatus = "Investigating"
		database.DB.Model(&incident).Updates(map[string]interface{}{
			"status":      newStatus,
			"assignee_id": uid,
		})
	} else if actionType == "RESOLVE" {
		newStatus = "Resolved"
		database.DB.Model(&incident).Update("status", newStatus)

		// Đóng TẤT CẢ các cảnh báo (Alerts) trên Mongo
		if database.SecurityAlertCollection != nil {
			_, err := database.SecurityAlertCollection.UpdateMany(context.TODO(), bson.M{"incident_id": incident.ID}, bson.M{"$set": bson.M{"is_resolved": true}})
			if err != nil {
				fmt.Printf("Warning: Failed to resolve associated alerts for incident %d: %v\n", incident.ID, err)
			}
		}

		// HẠ ĐIỂM RỦI RO NGAY LẬP TỨC
		scoring.RecalculateRiskScore(incident.AgentHWID)
	}

	// --- LOGIC LƯU FILE ẢNH VÀO Ổ CỨNG ---
	var imagePaths []string
	form, err := c.MultipartForm()

	// Lấy OrgID của User đang thao tác để tạo thư mục riêng
	var currentUser models.User
	database.DB.First(&currentUser, uid)
	orgID := currentUser.OrgID

	if err == nil && form != nil {
		files := form.File["images"]

		// 1. CHIA THƯ MỤC THEO CÔNG TY: uploads/org_1/incidents/
		uploadDir := fmt.Sprintf("uploads/org_%d/incidents", *orgID)
		os.MkdirAll(uploadDir, os.ModePerm)

		for _, file := range files {
			// 2. KIỂM TRA ĐUÔI FILE (Extension)
			ext := strings.ToLower(filepath.Ext(file.Filename))
			allowedExts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".pdf": true}
			if !allowedExts[ext] {
				fmt.Println("Khóa file rác:", file.Filename)
				continue // Bỏ qua file không hợp lệ
			}

			// 3. KIỂM TRA MAGIC BYTES (MIME TYPE) CHỐNG ĐỔI ĐUÔI GIẢ MẠO
			openedFile, _ := file.Open()
			buffer := make([]byte, 512) // Đọc 512 byte đầu tiên
			openedFile.Read(buffer)
			openedFile.Close()

			mimeType := http.DetectContentType(buffer)
			if !strings.HasPrefix(mimeType, "image/") && mimeType != "application/pdf" {
				fmt.Println("Khóa file giả mạo MIME:", file.Filename)
				continue
			}

			// Lưu file an toàn
			filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
			filepathStr := filepath.Join(uploadDir, filename)

			if err := c.SaveUploadedFile(file, filepathStr); err == nil {
				// 4. CHỈ LƯU TÊN FILE VÀO DB (Không lưu nguyên đường dẫn thật)
				imagePaths = append(imagePaths, filename)
			}
		}
	}

	// Convert mảng đường dẫn ảnh thành JSON String
	imagesJSON, _ := json.Marshal(imagePaths)

	// Ensure incident.ID is valid before creating activity
	if incident.ID == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Incident ID is zero, cannot add activity"})
		return
	}

	// --- LƯU TIMELINE (NHẬT KÝ HOẠT ĐỘNG) ---
	activity := models.IncidentActivity{
		IncidentID: incident.ID,
		UserID:     &uid,
		ActionType: actionType,
		Content:    content,
		OldStatus:  oldStatus,
		NewStatus:  incident.Status,
		Images:     string(imagesJSON),
	}

	if err := database.DB.Create(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add incident activity: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Cập nhật thành công",
		"activity": activity,
	})
}

func GetIncidentImage(c *gin.Context) {
	filename := c.Param("filename")

	// Lấy User từ Context (Do Middleware AuthRequired đã nạp vào)
	userIDVal, _ := c.Get("userID")
	var user models.User
	if err := database.DB.First(&user, userIDVal).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không xác định được danh tính"})
		return
	}

	// Trỏ tới đúng thư mục của công ty User đó
	filePath := fmt.Sprintf("uploads/org_%d/incidents/%s", *user.OrgID, filename)

	// Kiểm tra file có tồn tại không
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy file hoặc bạn không có quyền xem!"})
		return
	}

	// Trả file về cho trình duyệt
	c.File(filePath)
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
	incidentID := parseUint(id)
	if incidentID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}
	database.DB.Model(&models.Incident{}).Where("id = ?", incidentID).Update("playbook_progress", string(jsonBytes))
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

	incidentID := parseUint(id)
	if incidentID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	// Use IncidentService to update status and assignee
	if err := database.DB.Model(&models.Incident{}).Where("id = ?", incidentID).Updates(map[string]interface{}{"status": "Investigating", "assignee_id": userID}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign incident: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Đã nhận xử lý sự cố"})
}

// Helper function to parse uint from string
func parseUint(s string) uint {
	var i uint
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return 0
	}
	return i
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
