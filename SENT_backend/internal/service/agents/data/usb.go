package data

import (
	"context"
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/incidents"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	if database.USBCollection == nil {
		return
	}

	incSvc := &incidents.IncidentService{}

	incomingUsbMap := make(map[string]bool)

	for _, rec := range records {
		incomingUsbMap[rec.DeviceHash] = true
		filter := bson.M{"agent_hwid": agent.HWID, "device_hash": rec.DeviceHash}
		update := bson.M{"$set": bson.M{
			"agent_hwid":    agent.HWID,
			"device_name":   rec.DeviceName,
			"device_id":     rec.DeviceID,
			"vid":           rec.VID,
			"pid":           rec.PID,
			"serial_number": rec.SerialNumber,
			"device_hash":   rec.DeviceHash,
			"event_type":    "CONNECTED",
			"updated_at":    time.Now(),
		}}
		_, _ = database.USBCollection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))

		incSvc.TriggerSecurityEvent(agent,
			"USB Violation",
			"[P3] Thiết bị ngoại vi mới",
			fmt.Sprintf("Phát hiện USB lạ: %s (VID: %s)", rec.DeviceName, rec.VID),
			"P3",
		)
	}

	cursor, err := database.USBCollection.Find(context.TODO(), bson.M{"agent_hwid": agent.HWID, "event_type": "CONNECTED"})
	if err != nil {
		return
	}
	var connectedUSBs []models.USBLog
	cursor.All(context.TODO(), &connectedUSBs)

	for _, dbUsb := range connectedUSBs {
		if !incomingUsbMap[dbUsb.DeviceHash] {
			_, _ = database.USBCollection.UpdateOne(context.TODO(), bson.M{"_id": dbUsb.ID}, bson.M{"$set": bson.M{"event_type": "DISCONNECTED"}})
		}
	}
}
