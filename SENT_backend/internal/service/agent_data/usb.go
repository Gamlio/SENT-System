package agent_data

import (
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/security"
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

	for _, rec := range records {
		var existing models.USBLog
		// Kiểm tra xem USB Hash này đã từng cắm vào máy này chưa
		err := database.DB.Where("agent_hw_id = ? AND device_hash = ?", agent.HWID, rec.DeviceHash).First(&existing).Error

		if err != nil {
			// CHƯA TỪNG CẮM -> Ghi log và Báo động
			newLog := models.USBLog{
				AgentHWID:    agent.HWID,
				DeviceName:   rec.DeviceName,
				DeviceID:     rec.DeviceID,
				VID:          rec.VID,
				PID:          rec.PID,
				SerialNumber: rec.SerialNumber,
				DeviceHash:   rec.DeviceHash,
				EventType:    "active",
			}
			database.DB.Create(&newLog)

			security.TriggerSecurityEvent(agent,
				"USB Violation",
				"[P3] Cắm thiết bị USB chưa xác thực",
				fmt.Sprintf("Phát hiện thiết bị ngoại vi lạ: %s (VID: %s, PID: %s).", rec.DeviceName, rec.VID, rec.PID),
				"P3",
			)
		}
	}

	go security.CheckUSBCompliance(agent, data)
}
