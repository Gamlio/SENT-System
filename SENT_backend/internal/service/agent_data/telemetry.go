package agent_data

import (
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"time"

	"gorm.io/gorm"
)

// ProcessTelemetry xử lý danh sách cổng mạng và cập nhật IP từ Payload
func ProcessTelemetry(agent models.Agent, data interface{}) {
	// 1. Ép kiểu dữ liệu để đọc được các trường bên trong (quan trọng!)
	telemetryData, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	// 2. Chuẩn bị dữ liệu cần cập nhật vào bảng Agents
	updateFields := map[string]interface{}{
		"status":    "online",
		"last_seen": time.Now(),
	}

	// [FIX QUAN TRỌNG] Kiểm tra xem có 'ip_address' gửi lên không
	// Nếu có thì lấy giá trị đó ghi đè vào DB (thay vì dùng IP kết nối ::1)
	if ip, exists := telemetryData["ip_address"]; exists && ip != "" {
		updateFields["ip_address"] = fmt.Sprintf("%v", ip)
	}

	// Thực hiện cập nhật vào Database
	database.DB.Model(&agent).Updates(updateFields)

	// ---------------------------------------------------------
	// Phần dưới này là logic xử lý Port (giữ nguyên như cũ)
	// ---------------------------------------------------------

	// 3. Tính toán Hash của dữ liệu mới để so sánh
	newHash := CalculateHash(data)

	// 4. Kiểm tra bản ghi Snapshot cũ trong DB
	var snapshot models.AgentSnapshot
	err := database.DB.Where("agent_hw_id = ?", agent.HWID).First(&snapshot).Error

	// Nếu Hash không đổi -> Dữ liệu cổng mạng y hệt lần trước, không cần ghi lại
	if err == nil && snapshot.LastPortHash == newHash {
		return
	}

	// 5. Nếu có thay đổi, tiến hành cập nhật danh sách cổng mạng
	var payload struct {
		OpenPorts []models.OpenPort `json:"open_ports"`
	}
	// Hàm MapToStruct này nằm trong helpers.go
	if err := MapToStruct(data, &payload); err != nil {
		return
	}

	// Sử dụng Transaction để đảm bảo xóa cũ - thêm mới an toàn
	database.DB.Transaction(func(tx *gorm.DB) error {
		// Xóa các cổng mạng cũ của máy này
		tx.Where("agent_hw_id = ?", agent.HWID).Delete(&models.OpenPort{})

		// Thêm danh sách cổng mạng mới
		for _, p := range payload.OpenPorts {
			p.AgentHWID = agent.HWID
			if err := tx.Create(&p).Error; err != nil {
				return err
			}
		}

		// Cập nhật lại mã Hash mới vào Snapshot
		snapshot.AgentHWID = agent.HWID
		snapshot.LastPortHash = newHash
		tx.Save(&snapshot)

		return nil
	})
}
