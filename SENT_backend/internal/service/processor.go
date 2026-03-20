package service

import (
	"log"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/agent_data"
	"time"
)

type AgentPayload struct {
	Type     string      `json:"type"` // "DATA" | "HEARTBEAT"
	LogType  string      `json:"log_type"`
	HWID     string      `json:"hwid"`
	Hostname string      `json:"hostname"`
	Data     interface{} `json:"data"`
}

func ProcessAgentData(payload AgentPayload) {
	// 1. Tối ưu: Chỉ cập nhật LastSeen và Hostname.
	// TUYỆT ĐỐI KHÔNG cập nhật "status": "online" để tránh ghi đè trạng thái Zero Trust (ACTIVE/PENDING)
	database.DB.Model(&models.Agent{}).
		Where("hw_id = ?", payload.HWID).
		Updates(map[string]interface{}{
			"last_seen": time.Now(),
			"hostname":  payload.Hostname,
		})

	// Nếu chỉ là nhịp đập tim để báo online thì dừng ở đây
	if payload.Type == "HEARTBEAT" {
		return
	}

	// 2. Lấy thông tin Agent để truyền vào các hàm xử lý
	var agent models.Agent
	if err := database.DB.Where("hw_id = ?", payload.HWID).First(&agent).Error; err != nil {
		log.Printf("❌ Lỗi: Không tìm thấy Agent %s", payload.HWID)
		return
	}

	// [LỚP BẢO VỆ THỨ 2]: Khóa dữ liệu rác
	// Nếu máy chưa được Duyệt (ACTIVE) thì không được phép nhét rác vào các bảng USB, Software...
	if agent.Status != "ACTIVE" {
		log.Printf("🛡️ Bỏ qua dữ liệu [%s] từ máy %s vì trạng thái đang là %s", payload.LogType, agent.HWID, agent.Status)
		return
	}

	// 3. Phân phối dữ liệu vào các module chuyên trách
	// Chú ý: Ta ném nguyên cục 'payload.Data' cho agent_data tự ép kiểu JSON, giúp file này cực kỳ gọn gàng.
	switch payload.LogType {
	case "software":
		agent_data.ProcessSoftware(agent, payload.Data)
	case "usb":
		agent_data.ProcessUSB(agent, payload.Data)
	case "network": // <--- CASE MỚI THAY THẾ TELEMETRY
		agent_data.ProcessNetwork(agent, payload.Data)
	case "ports": // <--- CASE MỚI DÀNH RIÊNG CHO PORT
		agent_data.ProcessPorts(agent, payload.Data)
	default:
		log.Printf("⚠️ Cảnh báo: Loại log không hợp lệ [%s] từ máy %s", payload.LogType, agent.Hostname)
	}
}
