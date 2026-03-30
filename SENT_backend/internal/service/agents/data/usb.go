package data

import (
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/incidents"
	"time"
)

type AgentUSBRecord struct {
	DeviceName   string `json:"device_name"`
	DeviceID     string `json:"device_id"`
	VID          string `json:"vid"`
	PID          string `json:"pid"`
	SerialNumber string `json:"serial_number"`
	DeviceHash   string `json:"device_hash"`
	EventType    string `json:"event_type"`
}

func ProcessUSB(agent models.Agent, data interface{}) {
	bytes, _ := json.Marshal(data)
	var records []AgentUSBRecord
	if err := json.Unmarshal(bytes, &records); err != nil {
		return
	}

	incSvc := &incidents.IncidentService{}

	// 1. TẠO MAP CHỨA DANH SÁCH USB ĐANG CẮM
	incomingUsbMap := make(map[string]bool)

	// 2. CHIỀU THÊM MỚI / CẬP NHẬT (UPSERT)
	for _, rec := range records {
		incomingUsbMap[rec.DeviceHash] = true

		var existingUSB models.USBLog
		result := database.DB.Where("agent_hw_id = ? AND device_hash = ?", agent.HWID, rec.DeviceHash).First(&existingUSB)

		if result.Error == nil {
			// USB đã từng cắm -> Cập nhật trạng thái thành CONNECTED và update thời gian
			database.DB.Model(&existingUSB).Updates(map[string]interface{}{
				"event_type": "CONNECTED",
				"updated_at": time.Now(),
			})
			continue
		}

		// USB Mới Tinh -> Lưu vào DB
		usb := models.USBLog{
			AgentHWID:    agent.HWID,
			DeviceName:   rec.DeviceName,
			DeviceID:     rec.DeviceID,
			VID:          rec.VID,
			PID:          rec.PID,
			SerialNumber: rec.SerialNumber,
			DeviceHash:   rec.DeviceHash,
			EventType:    "CONNECTED", // Ép trạng thái ban đầu là đang cắm
		}
		database.DB.Create(&usb)

		// Phát cảnh báo do có USB mới
		incSvc.TriggerSecurityEvent(agent,
			"USB Violation",
			"[P3] Thiết bị ngoại vi mới",
			fmt.Sprintf("Phát hiện USB lạ: %s (VID: %s)", rec.DeviceName, rec.VID),
			"P3",
		)
	}

	// 3. CHIỀU RÚT RA (DIFFING): Tìm các USB vừa bị rút
	var connectedUSBs []models.USBLog
	// Tìm các USB đang được ghi nhận là CONNECTED
	database.DB.Where("agent_hw_id = ? AND event_type = ?", agent.HWID, "CONNECTED").Find(&connectedUSBs)

	for _, dbUsb := range connectedUSBs {
		// Nếu USB trong DB KHÔNG có trong danh sách Agent gửi lên
		if !incomingUsbMap[dbUsb.DeviceHash] {
			// Đánh dấu là đã rút ra
			database.DB.Model(&dbUsb).Update("event_type", "DISCONNECTED")
		}
	}
}
