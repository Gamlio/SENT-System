package agents

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/security"
	"strings"

	"github.com/gin-gonic/gin"
)

// Cấu trúc Payload chuẩn từ Agent gửi lên
type AgentPayload struct {
	LogType     string          `json:"log_type"`
	CompanyCode string          `json:"company_code"`
	HWID        string          `json:"hwid"`
	Hostname    string          `json:"hostname"`
	Data        json.RawMessage `json:"data"`
}

// Struct hứng dữ liệu từ Agent (Phải khớp 100% với Agent)
type AgentSoftwareRecord struct {
	SoftwareName    string `json:"software_name"`
	Version         string `json:"version"`
	Publisher       string `json:"publisher"`
	InstallLocation string `json:"install_location"`
	FileHash        string `json:"file_hash"`
	Status          string `json:"status"`
	IsRunning       bool   `json:"is_running"`
}

type AgentUSBRecord struct {
	DeviceName   string `json:"device_name"`
	DeviceID     string `json:"device_id"`
	VID          string `json:"vid"`
	PID          string `json:"pid"`
	SerialNumber string `json:"serial_number"`
	DeviceHash   string `json:"device_hash"`
	EventType    string `json:"event_type"`
}

// PushDataHandler: Gateway tiếp nhận dữ liệu tốc độ cao
func PushDataHandler(c *gin.Context) {
	var req AgentPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	var agent models.Agent
	if err := database.DB.Where("hw_id = ? AND org_id IN (SELECT id FROM organizations WHERE company_code = ?)", req.HWID, req.CompanyCode).First(&agent).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Agent không hợp lệ"})
		return
	}

	// Đẩy vào hàng đợi xử lý ngầm (Goroutine)
	go ProcessAgentDataAsync(req, agent)

	c.JSON(http.StatusOK, gin.H{"message": "Data accepted", "status": "ACTIVE"})
}

// Xử lý ngầm
func ProcessAgentDataAsync(payload AgentPayload, agent models.Agent) {
	switch payload.LogType {
	case "software":
		var records []AgentSoftwareRecord
		if err := json.Unmarshal(payload.Data, &records); err == nil {
			AnalyzeSoftwareBehavior(records, agent)
		}
	case "usb":
		var records []AgentUSBRecord
		if err := json.Unmarshal(payload.Data, &records); err == nil {
			AnalyzeUSBBehavior(records, agent)
		}
	}
}

// Lưu Software và Check Hash
func AnalyzeSoftwareBehavior(records []AgentSoftwareRecord, agent models.Agent) {
	// Xóa log cũ để cập nhật list phần mềm mới nhất
	database.DB.Where("agent_hw_id = ?", agent.HWID).Delete(&models.SoftwareItem{})

	for _, rec := range records {
		// Lưu DB
		sw := models.SoftwareItem{
			AgentHWID:       agent.HWID,
			SoftwareName:    rec.SoftwareName,
			Version:         rec.Version,
			Publisher:       rec.Publisher,
			InstallLocation: rec.InstallLocation,
			FileHash:        rec.FileHash,
			Status:          rec.Status,
			IsRunning:       rec.IsRunning,
		}
		database.DB.Create(&sw)

		// --- CASE 1: MÃ ĐỘC (YARA/Threat Intel Match) ---
		if rec.FileHash != "" && CheckHashAgainstThreatIntel(rec.FileHash) {
			security.TriggerSecurityEvent(agent,
				"Malware Detected", // Exact Type
				"[P1] Cảnh báo Mã Độc (YARA Match)",
				fmt.Sprintf("Tiến trình '%s' chứa mã băm độc hại: %s", rec.SoftwareName, rec.FileHash),
				"P1",
			)
		}

		// --- CASE 2: LẨN TRÁNH (Ghost Registry) ---
		if rec.Status == "GHOST_REGISTRY" {
			security.TriggerSecurityEvent(agent,
				"Defense Evasion", // Kỹ thuật lẩn tránh
				"[P2] Xóa dấu vết phần mềm (Ghost Registry)",
				fmt.Sprintf("Phần mềm '%s' đã bị xóa file vật lý nhưng vẫn giữ lại Registry. Có dấu hiệu che giấu hành vi.", rec.SoftwareName),
				"P2",
			)
		}

		// --- CASE 3: PHẦN MỀM CẤM (Ví dụ: Torrent) ---
		if CheckBannedSoftware(rec.SoftwareName) {
			security.TriggerSecurityEvent(agent,
				"Software Violation",
				"[P3] Cài đặt phần mềm bị cấm",
				fmt.Sprintf("Phát hiện phần mềm vi phạm chính sách công ty: %s", rec.SoftwareName),
				"P3",
			)
		}
	}
}

// 2. PHÂN TÍCH USB (Bắt Case: USB lạ / BadUSB)
func AnalyzeUSBBehavior(records []AgentUSBRecord, agent models.Agent) {
	for _, rec := range records {
		var existing models.USBLog
		err := database.DB.Where("agent_hw_id = ? AND device_hash = ?", agent.HWID, rec.DeviceHash).First(&existing).Error

		if err != nil { // Lần đầu cắm USB này
			usb := models.USBLog{
				AgentHWID:  agent.HWID,
				DeviceName: rec.DeviceName,
				DeviceID:   rec.DeviceID,
				VID:        rec.VID,
				PID:        rec.PID,
				DeviceHash: rec.DeviceHash,
				EventType:  rec.EventType,
			}
			database.DB.Create(&usb)

			// --- CASE 4: USB LẠ CHƯA DUYỆT ---
			security.TriggerSecurityEvent(agent,
				"USB Violation",
				"[P3] Cắm thiết bị ngoại vi trái phép",
				fmt.Sprintf("Phát hiện USB lạ: %s (VID: %s, PID: %s). Yêu cầu xác thực phần cứng.", rec.DeviceName, rec.VID, rec.PID),
				"P3",
			)
		}
	}
}

// 3. PHÂN TÍCH TƯỜNG LỬA / MẠNG (Nếu bạn gửi log Telemetry lên)
func AnalyzeTelemetry(firewallOff bool, agent models.Agent) {
	// --- CASE 5: TẮT TƯỜNG LỬA ---
	if firewallOff {
		security.TriggerSecurityEvent(agent,
			"Firewall Disabled",
			"[P1] Tường lửa hệ thống bị vô hiệu hóa",
			"Tường lửa của hệ điều hành đã bị tắt. Máy trạm mất lớp khiên bảo vệ mạng.",
			"P1",
		)
	}
}

// --- Mock Checkers ---
func CheckHashAgainstThreatIntel(hash string) bool {
	// Giả lập hash của 1 con virus
	maliciousHashes := map[string]bool{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855": true}
	return maliciousHashes[hash]
}

func CheckBannedSoftware(name string) bool {
	// Khai báo mảng chứa các từ khóa phần mềm bị cấm
	banned := []string{"utorrent", "cheat engine", "idm"}

	// Chuyển tên phần mềm về chữ thường để dễ so sánh
	nameLower := strings.ToLower(name)

	// Sử dụng biến banned để kiểm tra
	for _, b := range banned {
		if strings.Contains(nameLower, b) {
			return true // Báo vi phạm nếu chứa từ khóa cấm
		}
	}

	return false
}
