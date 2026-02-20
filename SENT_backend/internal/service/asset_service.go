package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AgentPayload struct {
	Type     string      `json:"type"` // "DATA" | "HEARTBEAT"
	LogType  string      `json:"log_type"`
	HWID     string      `json:"hwid"`
	Hostname string      `json:"hostname"`
	Data     interface{} `json:"data"`
}

func ProcessAgentData(payload AgentPayload) {
	// 1. LUÔN LUÔN CẬP NHẬT ONLINE (Dù là Heartbeat hay Data)
	database.DB.Model(&models.Agent{}).
		Where("hw_id = ?", payload.HWID).
		Updates(map[string]interface{}{
			"status":    "online",
			"last_seen": time.Now(),
			"hostname":  payload.Hostname, // Cập nhật luôn tên máy nếu đổi
		})

	// 2. Nếu là Heartbeat -> Dừng luôn, không làm gì nữa (Tối ưu Server)
	if payload.Type == "HEARTBEAT" {
		return
	}

	// 3. Nếu là DATA -> Bắt đầu quy trình xử lý & AI
	if payload.Type == "DATA" {
		// Fetch Agent from database
		var agent models.Agent
		if err := database.DB.Where("hw_id = ?", payload.HWID).First(&agent).Error; err != nil {
			log.Printf("Agent not found: %v", err)
			return
		}

		switch payload.LogType {
		case "telemetry":
			// Lưu dữ liệu vào DB (Code cũ)
			ProcessTelemetry(agent, payload.Data)

			// --- KÍCH HOẠT AI SECURITY ---
			go AnalyzeBehaviorAI(payload.HWID, "telemetry", payload.Data)

		case "software":
			ProcessSoftware(agent, payload.Data)

			// --- KÍCH HOẠT AI COMPLIANCE ---
			go AnalyzeBehaviorAI(payload.HWID, "software", payload.Data)
		}
	}
}
func AnalyzeBehaviorAI(hwid string, dataType string, data interface{}) {
	// Đây là nơi bạn sẽ cắm mô hình AI vào sau này
	// Ví dụ logic đơn giản:

	log.Printf("🤖 AI đang phân tích hành vi mới của máy %s...", hwid)

	// Ví dụ: Phát hiện mở cổng lạ
	if dataType == "telemetry" {
		// Logic AI: Nếu thấy cổng 22 (SSH) hoặc 3389 (RDP) mở bất thường vào ban đêm -> Báo động
		// createAlert(hwid, "AI_ANOMALY", "Phát hiện hành vi mở cổng điều khiển từ xa đáng ngờ")
	}
}

// --- Helper: Tính toán Hash của dữ liệu JSON ---
func calculateHash(data interface{}) string {
	bytes, _ := json.Marshal(data)
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

// 1. Xử lý Inventory (Phần cứng)
func ProcessInventory(agent models.Agent, data interface{}) {
	invData, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	var inv models.AgentInventory
	// Tìm xem máy này đã có cấu hình trong DB chưa
	result := database.DB.Where("hw_id = ?", agent.HWID).First(&inv)

	inv.AgentHWID = agent.HWID
	inv.OSInfo = fmt.Sprintf("%v", invData["os_info"])
	inv.CPUModel = fmt.Sprintf("%v", invData["cpu_model"])

	// Convert Float64 sang Int một cách an toàn
	if ramFloat, ok := invData["ram_total_gb"].(float64); ok {
		inv.RAMTotalGB = int(ramFloat)
	}

	if result.Error != nil {
		database.DB.Create(&inv) // Máy mới -> Tạo mới
	} else {
		database.DB.Save(&inv) // Máy cũ -> Cập nhật
	}
}

// 2. Xử lý Telemetry (Port, USB) - CÓ SỬ DỤNG SNAPSHOT ĐỂ TỐI ƯU
func ProcessTelemetry(agent models.Agent, data interface{}) {
	// Cập nhật trạng thái "Online" cho Agent
	database.DB.Model(&agent).Updates(map[string]interface{}{
		"status":    "online",
		"last_seen": time.Now(),
	})

	// Tính Hash dữ liệu mới
	newHash := calculateHash(data)

	// Kiểm tra Snapshot cũ
	var snapshot models.AgentSnapshot
	err := database.DB.Where("agent_hw_id = ?", agent.HWID).First(&snapshot).Error

	// NẾU HASH GIỐNG NHAU -> DỪNG LẠI (TIẾT KIỆM DB)
	if err == nil && snapshot.LastPortHash == newHash {
		// Log nhẹ để biết là đã bỏ qua (khi Dev), Production thì comment lại
		// log.Println("⚡ Telemetry không đổi, bỏ qua update DB.")
		return
	}

	// Nếu khác, bắt đầu xử lý
	var payload struct {
		OpenPorts []models.OpenPort `json:"open_ports"`
	}
	if err := mapToStruct(data, &payload); err != nil {
		return
	}

	// Transaction để đảm bảo toàn vẹn dữ liệu
	database.DB.Transaction(func(tx *gorm.DB) error {
		// Xóa dữ liệu cũ
		tx.Where("agent_hw_id = ?", agent.HWID).Delete(&models.OpenPort{})

		// Thêm dữ liệu mới
		for _, p := range payload.OpenPorts {
			p.AgentHWID = agent.HWID
			tx.Create(&p)
		}

		// Cập nhật Snapshot Hash mới
		snapshot.AgentHWID = agent.HWID
		snapshot.LastPortHash = newHash
		tx.Save(&snapshot)
		return nil
	})
}

// ProcessSoftware: Xử lý log phần mềm từ Agent gửi lên
func ProcessSoftware(agent models.Agent, data interface{}) {
	softwareList, ok := data.([]interface{})
	if !ok {
		return
	}

	// Xóa danh sách cũ để cập nhật mới
	database.DB.Where("hw_id = ?", agent.HWID).Delete(&models.SoftwareItem{})

	// 1. LẤY WHITELIST CỦA RIÊNG MÁY NÀY
	var localWhitelist []models.AgentWhitelist
	database.DB.Where("hwid = ?", agent.HWID).Find(&localWhitelist)
	isZeroTrustMode := len(localWhitelist) > 0 // Kích hoạt Zero Trust nếu có dữ liệu

	// 2. LẤY BLACKLIST CHUNG CỦA HỆ THỐNG
	var globalBlacklist []models.UniversalPolicy
	database.DB.Where("policy_type = ?", "blacklist").Find(&globalBlacklist)

	for _, sw := range softwareList {
		swName := fmt.Sprintf("%v", sw)
		swNameLower := strings.ToLower(swName)

		// Lưu phần mềm vào DB để hiển thị
		dbItem := models.SoftwareItem{AgentHWID: agent.HWID, SoftwareName: swName}
		database.DB.Create(&dbItem)

		// --- LOGIC KIỂM TRA CHÉO ---
		if isZeroTrustMode {
			// CHẾ ĐỘ 1: ZERO TRUST (Cực ngặt)
			// Tất cả phải nằm trong Whitelist, nếu trật ra ngoài -> Báo động
			isAllowed := false
			for _, white := range localWhitelist {
				if strings.Contains(swNameLower, strings.ToLower(white.SoftwareName)) {
					isAllowed = true
					break
				}
			}
			if !isAllowed {
				alert := models.SecurityAlert{
					HWID: agent.HWID, OrgID: agent.OrgID, AlertType: "Vi phạm Zero Trust", Severity: "Critical",
					Description: fmt.Sprintf("Phần mềm KHÔNG nằm trong Whitelist: %s", swName),
				}
				database.DB.Create(&alert)
			}
		} else {
			// CHẾ ĐỘ 2: KIỂM TRA BLACKLIST (Chế độ thường)
			for _, black := range globalBlacklist {
				if black.PolicyType != "" && strings.Contains(swNameLower, strings.ToLower(black.PolicyType)) {
					alert := models.SecurityAlert{
						HWID: agent.HWID, OrgID: agent.OrgID, AlertType: "Phần mềm trái phép", Severity: "High",
						Description: fmt.Sprintf("Phát hiện phần mềm bị cấm chung: %s", swName),
					}
					database.DB.Create(&alert)
					break
				}
			}
		}
	}
}

// Hàm tạo cảnh báo nhanh
func createAlert(tx *gorm.DB, agent models.Agent, alertType, desc string) {
	// Kiểm tra xem đã có cảnh báo này chưa (tránh spam alert)
	var count int64
	tx.Model(&models.SecurityAlert{}).Where("hw_id = ? AND alert_type = ? AND is_resolved = ?",
		agent.HWID, alertType, false).Count(&count)

	if count == 0 {
		tx.Create(&models.SecurityAlert{
			OrgID:       agent.OrgID,
			HWID:        agent.HWID,
			AlertType:   alertType,
			Title:       "Phát hiện vi phạm tuân thủ",
			Description: desc,
			Severity:    "High",
		})
	}
}

func mapToStruct(input interface{}, output interface{}) error {
	b, _ := json.Marshal(input)
	return json.Unmarshal(b, output)
}
