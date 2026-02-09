package service

import (
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"time"
)

// --- CÁC STRUCT HỨNG DỮ LIỆU JSON TỪ AGENT V3.1 ---
// (Dùng để ép kiểu dữ liệu interface{} sang struct có nghĩa)

type InventoryPayload struct {
	CPUModel   string `json:"cpu_model"`
	RAMTotalGB int    `json:"ram_total_gb"`
	OSInfo     string `json:"os_info"`
}

type TelemetryPayload struct {
	RAMUsedPercent float64               `json:"ram_used_percent"`
	OpenPorts      []models.OpenPort     `json:"open_ports"`
	USBDevices     []models.USBWhitelist `json:"usb_devices"` // Agent gửi về name & id
}

// --- LOGIC XỬ LÝ CHÍNH ---

// 1. Xử lý Inventory (Thông tin phần cứng)
func ProcessInventory(agent models.Agent, data interface{}) {
	var payload InventoryPayload
	if err := mapToStruct(data, &payload); err != nil {
		fmt.Println("❌ Lỗi parse Inventory:", err)
		return
	}

	// Cập nhật hoặc tạo mới thông tin phần cứng
	inventory := models.AgentInventory{
		AgentHWID:  agent.HWID,
		CPUModel:   payload.CPUModel,
		RAMTotalGB: payload.RAMTotalGB,
		OSInfo:     payload.OSInfo,
	}

	// Sử dụng Save để Insert hoặc Update
	var existing models.AgentInventory
	if err := database.DB.Where("agent_hwid = ?", agent.HWID).First(&existing).Error; err == nil {
		inventory.ID = existing.ID // Giữ ID cũ để update
	}
	database.DB.Save(&inventory)
}

// 2. Xử lý Software (Kiểm tra tuân thủ phần mềm)
func ProcessSoftware(agent models.Agent, data interface{}) {
	// Dữ liệu Software từ Agent là mảng string []string
	softwareList, ok := data.([]interface{})
	if !ok {
		return
	}

	// Xóa danh sách cũ (Làm mới Inventory)
	database.DB.Where("agent_hwid = ?", agent.HWID).Delete(&models.SoftwareItem{})

	var violations []string

	for _, item := range softwareList {
		name := item.(string)

		// A. Lưu vào DB
		database.DB.Create(&models.SoftwareItem{
			AgentHWID:    agent.HWID,
			SoftwareName: name,
			Version:      "Detected",
		})

		// B. CHECK COMPLIANCE: Kiểm tra xem có nằm trong danh sách cấm không?
		var policy models.SoftwarePolicy
		// Tìm xem công ty này có cấm phần mềm này không
		err := database.DB.Where("org_id = ? AND software_name = ? AND is_prohibited = ?", agent.OrgID, name, true).First(&policy).Error
		if err == nil {
			violations = append(violations, name)
		}
	}

	// C. Tạo Cảnh báo nếu có vi phạm
	if len(violations) > 0 {
		CreateAlert(agent, "SOFTWARE_VIOLATION", "High",
			fmt.Sprintf("Phát hiện %d phần mềm bị cấm: %v", len(violations), violations))
	}
}

// 3. Xử lý Telemetry (RAM, Port, USB - QUAN TRỌNG NHẤT)
func ProcessTelemetry(agent models.Agent, data interface{}) {
	var payload TelemetryPayload
	if err := mapToStruct(data, &payload); err != nil {
		return
	}

	// A. Cập nhật trạng thái RAM & LastSeen
	database.DB.Model(&agent).Updates(map[string]interface{}{
		"status":    "online",
		"last_seen": time.Now(),
	})

	// B. Lưu Open Ports (Xóa cũ nạp mới cho realtime)
	database.DB.Where("agent_hwid = ?", agent.HWID).Delete(&models.OpenPort{})
	for _, p := range payload.OpenPorts {
		p.AgentHWID = agent.HWID
		database.DB.Create(&p)
	}

	// C. XỬ LÝ USB & WHITELIST (TRUY CỨU TRÁCH NHIỆM)
	for _, usb := range payload.USBDevices {
		// Bỏ qua các thiết bị hệ thống (Root Hub) để đỡ rác log
		if usb.DeviceID == "" {
			continue
		}

		// 1. Kiểm tra Whitelist
		var whitelist models.USBWhitelist
		var isAllowed bool = false

		// Tìm xem USB này (DeviceID) có được phép trong Org này không
		err := database.DB.Where("org_id = ? AND device_id = ?", agent.OrgID, usb.DeviceID).First(&whitelist).Error

		if err == nil {
			isAllowed = true
		} else {
			// 2. Nếu KHÔNG tìm thấy -> BẮN CẢNH BÁO NGAY!
			CreateAlert(agent, "USB_UNAUTHORIZED", "Critical",
				fmt.Sprintf("Phát hiện USB lạ: %s (%s)", usb.FriendlyName, usb.DeviceID))
		}

		// 3. Ghi Log lịch sử cắm
		database.DB.Create(&models.USBLog{
			AgentHWID:     agent.HWID,
			DeviceName:    usb.FriendlyName, // Agent gửi name vào field này
			DeviceID:      usb.DeviceID,
			IsWhitelisted: isAllowed,
			EventType:     "plugged",
		})
	}
}

// --- HÀM BỔ TRỢ ---

func CreateAlert(agent models.Agent, alertType, severity, desc string) {
	// Kiểm tra xem đã có alert chưa xử lý chưa để tránh spam DB
	var exists int64
	database.DB.Model(&models.SecurityAlert{}).Where(
		"hwid = ? AND alert_type = ? AND is_resolved = ?",
		agent.HWID, alertType, false).Count(&exists)

	if exists == 0 {
		alert := models.SecurityAlert{
			OrgID:       agent.OrgID,
			HWID:        agent.HWID,
			AlertType:   alertType,
			Title:       fmt.Sprintf("Cảnh báo an ninh tại máy %s", agent.Hostname),
			Description: desc,
			Severity:    severity,
		}
		database.DB.Create(&alert)
		fmt.Printf("🚨 ALERT CREATED: %s - %s\n", alertType, desc)
	}
}

// Hàm ép kiểu JSON map sang Struct
func mapToStruct(input interface{}, output interface{}) error {
	bytes, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, output)
}
