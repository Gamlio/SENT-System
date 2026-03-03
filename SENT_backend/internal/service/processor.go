package service

import (
	"log"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/agent_data" // Đảm bảo đúng đường dẫn từ go.mod
	"sent_backend/internal/service/security"
	"time"
)

// --- THÊM ĐOẠN NÀY VÀO ---
type AgentPayload struct {
	Type     string      `json:"type"` // "DATA" | "HEARTBEAT"
	LogType  string      `json:"log_type"`
	HWID     string      `json:"hwid"`
	Hostname string      `json:"hostname"`
	Data     interface{} `json:"data"`
}

// -------------------------
func ProcessAgentData(payload AgentPayload) {
	// 1. Cập nhật trạng thái Online ngay lập tức
	database.DB.Model(&models.Agent{}).
		Where("hw_id = ?", payload.HWID).
		Updates(map[string]interface{}{
			"status":    "online",
			"last_seen": time.Now(),
			"hostname":  payload.Hostname,
		})

	if payload.Type == "HEARTBEAT" {
		return
	}

	// 2. Lấy thông tin Agent để truyền vào các hàm xử lý
	var agent models.Agent
	if err := database.DB.Where("hw_id = ?", payload.HWID).First(&agent).Error; err != nil {
		log.Printf("Lỗi: Không tìm thấy Agent %s", payload.HWID)
		return
	}

	// 3. Phân phối dữ liệu vào các folder con
	switch payload.LogType {
	case "inventory":
		agent_data.ProcessInventory(agent, payload.Data)
	case "software":
		agent_data.ProcessSoftware(agent, payload.Data)
		// Gọi hàm từ package security
		go security.AnalyzeBehaviorAI(agent.HWID, "software", payload.Data)
	case "usb":
		agent_data.ProcessUSB(agent, payload.Data)
		go security.AnalyzeBehaviorAI(agent.HWID, "usb", payload.Data)
	case "telemetry":
		agent_data.ProcessTelemetry(agent, payload.Data)
		go security.AnalyzeBehaviorAI(agent.HWID, "telemetry", payload.Data)
	}
}
