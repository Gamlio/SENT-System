package data

import (
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/incidents"

	"gorm.io/gorm/clause"
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
		usb := models.USBLog{
			AgentHWID:    agent.HWID,
			DeviceName:   rec.DeviceName,
			DeviceID:     rec.DeviceID,
			VID:          rec.VID,
			PID:          rec.PID,
			SerialNumber: rec.SerialNumber,
			DeviceHash:   rec.DeviceHash,
			EventType:    rec.EventType,
		}

		// Upsert: Cập nhật EventType/Time nếu đã tồn tại, Tạo mới nếu chưa có
		result := database.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "agent_hw_id"}, {Name: "device_hash"}},
			DoUpdates: clause.AssignmentColumns([]string{"event_type", "updated_at"}),
		}).Create(&usb)

		if result.Error == nil {
			incSvc := &incidents.IncidentService{}
			isNew := usb.CreatedAt.Unix() == usb.UpdatedAt.Unix()
			if isNew {
				incSvc.TriggerSecurityEvent(agent,
					"USB Violation",
					"[P3] Thiết bị ngoại vi mới",
					fmt.Sprintf("Phát hiện USB lạ: %s (VID: %s)", rec.DeviceName, rec.VID),
					"P3",
				)
			}
		}
	}
}
