package agents

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Hàm random tạo chuỗi ngẫu nhiên bảo mật cao
func generateSecureToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// 1. LÀM CHO ADMIN SOC: Sinh mã cài đặt có hạn 24h
// POST /api/v1/agents/generate-token (Yêu cầu JWT)
func GenerateEnrollmentToken(c *gin.Context) {
	// Lấy OrgID từ JWT Middleware
	orgIDVal, _ := c.Get("org_id")
	usernameVal, _ := c.Get("username")

	// Lấy OrgID ép về kiểu uint
	orgID := orgIDVal.(uint)

	// [MỚI] Format mã cài đặt: SENT - [ID CÔNG TY] - [CHUỖI NGẪU NHIÊN 8 KÝ TỰ]
	// Ví dụ: SENT-01-A1B2C3D4
	randomPart := strings.ToUpper(generateSecureToken(4)) // 4 byte = 8 ký tự Hex
	tokenString := fmt.Sprintf("%02d-%s", orgID, randomPart)

	token := models.EnrollmentToken{
		Token: tokenString,
		OrgID: orgID,
		// [QUAN TRỌNG]: Giảm thời hạn xuống đúng 15 phút
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedBy: usernameVal.(string),
	}

	database.DB.Create(&token)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Tạo mã cài đặt thành công",
		"enroll_token": tokenString,
		"expires_at":   token.ExpiresAt,
	})
}

// 2. LÀM CHO PHẦN MỀM SENT: Đăng ký máy trạm mới
// POST /api/v1/agents/enroll (Public - Không cần JWT)
func EnrollAgent(c *gin.Context) {
	var req struct {
		HWID        string `json:"hwid"`
		Hostname    string `json:"hostname"`
		IPAddress   string `json:"ip_address"`
		EnrollToken string `json:"enroll_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Kiểm tra Token có tồn tại và còn hạn không?
	var tokenRecord models.EnrollmentToken
	if err := database.DB.Where("token = ?", req.EnrollToken).First(&tokenRecord).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Mã cài đặt không hợp lệ hoặc không tồn tại!"})
		return
	}

	if time.Now().After(tokenRecord.ExpiresAt) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Mã cài đặt đã hết hạn!"})
		return
	}

	secretKey := generateSecureToken(16)
	var existingAgent models.Agent
	result := database.DB.Where("hw_id = ?", req.HWID).First(&existingAgent)

	if result.Error == nil {
		// TRƯỜNG HỢP: ĐÃ TỒN TẠI[cite: 38]
		database.DB.Model(&existingAgent).Updates(map[string]interface{}{
			"status":     "PENDING",
			"secret_key": secretKey,
			"hostname":   req.Hostname,
			"ip_address": req.IPAddress,
			"last_seen":  time.Now(),
		})

		alert := models.SecurityAlert{
			OrgID:       existingAgent.OrgID,
			HWID:        existingAgent.HWID,
			Priority:    "P2",
			AlertType:   "Re-Enrollment Detected",
			Description: "Thiết bị vừa xin cấp lại khóa xác thực. Có thể do cài đặt lại HĐH hoặc bị giả mạo (Spoofing). Trạng thái đã bị đẩy về PENDING. Vui lòng xác minh trước khi phê duyệt.",
		}
		database.DB.Create(&alert)

	} else {
		// TRƯỜNG HỢP: MÁY MỚI HOÀN TOÀN[cite: 38]
		newAgent := models.Agent{
			HWID:      req.HWID,
			Hostname:  req.Hostname,
			IPAddress: req.IPAddress,
			OrgID:     tokenRecord.OrgID,
			Status:    "PENDING",
			SecretKey: secretKey,
			LastSeen:  time.Now(),
		}
		database.DB.Create(&newAgent)
	}

	// --- BỔ SUNG LỚP KEO DÍNH: TẠO APPROVAL TICKET ---
	// Kiểm tra xem máy này đã có Ticket nào đang PENDING chưa để tránh spam vé
	var existingTicket models.ApprovalTicket
	ticketExists := database.DB.Where("module_type = ? AND target_id = ? AND status = ?", "AGENT_ENROLL", req.HWID, "PENDING").First(&existingTicket)

	if ticketExists.Error != nil {
		// Chưa có Ticket nào chờ duyệt, tạo mới
		ticket := models.ApprovalTicket{
			OrgID:        tokenRecord.OrgID,
			ModuleType:   "AGENT_ENROLL",
			ActionType:   "ENROLL",
			TargetID:     0,
			TargetName:   req.HWID, // Ví dụ: PC-KETOAN-01 (WIN-1234)
			Status:       "PENDING",
			RequestedBy:  "SYSTEM", // Hệ thống tự động tạo do Agent yêu cầu
			SnapshotData: fmt.Sprintf(`{"ip": "%s", "action": "Yêu cầu kết nối vào SOC"}`, req.IPAddress),
		}
		database.DB.Create(&ticket)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Đăng ký thành công. Đang chờ SOC phê duyệt.",
		"secret_key": secretKey,
		"status":     "PENDING",
	})
}
