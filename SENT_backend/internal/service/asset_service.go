package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
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
	var payload models.AgentInventory
	if err := mapToStruct(data, &payload); err != nil {
		return
	}

	// Logic cập nhật thông thường (Inventory ít thay đổi, có thể update thẳng)
	payload.AgentHWID = agent.HWID

	// Tìm xem đã có chưa để lấy ID (tránh tạo mới liên tục)
	var existing models.AgentInventory
	if err := database.DB.Where("agent_hw_id = ?", agent.HWID).First(&existing).Error; err == nil {
		payload.ID = existing.ID
		payload.CreatedAt = existing.CreatedAt
	}

	database.DB.Save(&payload)
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

// 3. Xử lý Software - CÓ TỐI ƯU & CHẶN LOG RÁC
func ProcessSoftware(agent models.Agent, data interface{}) {
	softwareList, ok := data.([]interface{})
	if !ok {
		return
	}

	// 1. TỐI ƯU: So sánh Hash trước
	newHash := calculateHash(softwareList)
	var snapshot models.AgentSnapshot
	err := database.DB.Where("agent_hw_id = ?", agent.HWID).First(&snapshot).Error

	if err == nil && snapshot.LastSoftwareHash == newHash {
		// log.Println("⚡ Software không đổi, bỏ qua update DB.")
		return
	}

	// 2. Nếu có thay đổi -> Xử lý
	database.DB.Transaction(func(tx *gorm.DB) error {
		// Xóa danh sách cũ
		tx.Where("agent_hw_id = ?", agent.HWID).Delete(&models.SoftwareItem{})

		for _, item := range softwareList {
			name, _ := item.(string) // Giả sử Agent gửi mảng string tên phần mềm

			// Lưu vào SoftwareItem
			tx.Create(&models.SoftwareItem{
				AgentHWID:    agent.HWID,
				SoftwareName: name,
				Version:      "Detected",
			})

			// --- KIỂM TRA CHÍNH SÁCH (Fix log rác ở đây) ---
			var policy models.SoftwarePolicy
			// Tìm xem phần mềm này có bị cấm không
			err := tx.Where("org_id = ? AND software_name = ? AND is_prohibited = ?",
				agent.OrgID, name, true).First(&policy).Error

			// CHỈ LOG NẾU TÌM THẤY (LÀ CÓ VI PHẠM) HOẶC LỖI DB THỰC SỰ
			if err == nil {
				// Tìm thấy Policy cấm -> Tạo Alert
				createAlert(tx, agent, "SOFTWARE_VIOLATION", "Phần mềm bị cấm: "+name)
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				// Nếu lỗi KHÁC lỗi "Not Found" thì mới in ra console
				log.Printf("Lỗi DB khi check policy: %v", err)
			}
			// Nếu err == RecordNotFound -> Nghĩa là phần mềm sạch, không làm gì cả.
		}

		// Cập nhật Hash mới vào Snapshot
		snapshot.AgentHWID = agent.HWID
		snapshot.LastSoftwareHash = newHash
		tx.Save(&snapshot)
		return nil
	})
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
