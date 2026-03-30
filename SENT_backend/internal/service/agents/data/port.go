package data // Đổi sang package data

import (
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/incidents"
	"time"
)

type AgentTelemetryRecord struct {
	IPAddress   string            `json:"ip_address"`
	FirewallOff bool              `json:"firewall_off"`
	OpenPorts   []models.OpenPort `json:"open_ports"`
}

func ProcessPorts(agent models.Agent, data interface{}) {
	bytes, _ := json.Marshal(data)
	var payload AgentTelemetryRecord
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return
	}

	incSvc := &incidents.IncidentService{}

	// 1. TẠO MAP ĐỂ TRA CỨU NHANH DANH SÁCH PORT ĐANG GỬI LÊN
	incomingPortsMap := make(map[int]bool)

	// 2. XỬ LÝ CHIỀU THÊM MỚI / CẬP NHẬT (UPSERT)
	for _, incomingPort := range payload.OpenPorts {
		incomingPortsMap[incomingPort.Port] = true // Đánh dấu là có mặt

		var existingPort models.OpenPort
		result := database.DB.Where("agent_hw_id = ? AND port = ?", agent.HWID, incomingPort.Port).First(&existingPort)

		if result.Error == nil {
			// Port này đã tồn tại -> Đảm bảo nó đang ở trạng thái OPEN và cập nhật thời gian
			database.DB.Model(&existingPort).Updates(map[string]interface{}{
				"status":     "OPEN",
				"updated_at": time.Now(),
			})
			continue
		}

		// Port Mới Tinh -> Lưu vào DB
		incomingPort.AgentHWID = agent.HWID
		incomingPort.Status = "OPEN"
		database.DB.Create(&incomingPort)

		// CHỈ BÁO ĐỘNG KHI CÓ PORT NGUY HIỂM MỚI MỞ
		if incomingPort.Port == 3389 || incomingPort.Port == 22 || incomingPort.Port == 4444 {
			incSvc.TriggerSecurityEvent(agent, "Unauthorized Port",
				fmt.Sprintf("[P2] Mở cổng quản trị (%d) trái phép", incomingPort.Port),
				fmt.Sprintf("Tiến trình '%s' đang mở cổng %d.", incomingPort.ProcessName, incomingPort.Port), "P2")
		}
	}

	// 3. CHIỀU ĐÓNG (DIFFING): Tìm các Port cũ trong DB bị mất tích
	var activeDBPorts []models.OpenPort
	// Lấy tất cả các port của máy này đang được đánh dấu là OPEN trong DB
	database.DB.Where("agent_hw_id = ? AND status = ?", agent.HWID, "OPEN").Find(&activeDBPorts)

	for _, dbPort := range activeDBPorts {
		// Nếu Port trong DB KHÔNG có mặt trong danh sách Agent vừa gửi lên
		if !incomingPortsMap[dbPort.Port] {
			// Đánh dấu là đã đóng thay vì xóa data
			database.DB.Model(&dbPort).Update("status", "CLOSED")

			// Tùy chọn: Có thể ghi log hệ thống "Cổng X đã được đóng an toàn"
		}
	}
}
