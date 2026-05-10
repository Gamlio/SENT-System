package data

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type assetUSBRecord struct {
	DeviceName   string `json:"device_name"`
	DeviceID     string `json:"device_id"`
	VID          string `json:"vid"`
	PID          string `json:"pid"`
	SerialNumber string `json:"serial_number"`
	DeviceHash   string `json:"device_hash"`
	EventType    string `json:"event_type"`
}

func ProcessUSB(asset models.Asset, data interface{}) error {
	// Đổi tên biến thành bytesData để tránh trùng tên với package bytes
	bytesData, err := GetBytesFromData(data)
	if err != nil {
		return fmt.Errorf("dữ liệu USB không hợp lệ: %w", err)
	}

	log.Printf("[DEBUG] HWID: %s | Records received", asset.AssetHWID)

	var records []assetUSBRecord
	if err := json.Unmarshal(bytesData, &records); err != nil {
		return fmt.Errorf("lỗi giải mã JSON USB: %w", err)
	}

	if database.USBCollection == nil {
		return fmt.Errorf("database USBCollection chưa sẵn sàng")
	}

	var activeHashes []string

	for _, rec := range records {
		activeHashes = append(activeHashes, rec.DeviceHash)

		filter := bson.M{"asset_hwid": asset.AssetHWID, "org_id": int64(asset.OrgID), "device_hash": rec.DeviceHash}
		update := bson.M{"$set": bson.M{
			"asset_hwid":    asset.AssetHWID,
			"org_id":        int64(asset.OrgID),
			"device_name":   rec.DeviceName,
			"device_id":     rec.DeviceID,
			"vid":           rec.VID,
			"pid":           rec.PID,
			"serial_number": rec.SerialNumber,
			"device_hash":   rec.DeviceHash,
			"event_type":    "CONNECTED",
			"updated_at":    time.Now(),
		}}

		if _, err := database.USBCollection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true)); err != nil {
			return fmt.Errorf("lỗi lưu CSDL MongoDB USB: %w", err)
		}

		go func(r assetUSBRecord) {
			payload := map[string]interface{}{
				"asset":       asset,
				"alert_type":  "USB Violation",
				"title":       "[P3] Thiết bị ngoại vi mới",
				"description": fmt.Sprintf("Phát hiện USB lạ: %s (VID: %s, Serial: %s)", r.DeviceName, r.VID, r.SerialNumber),
				"priority":    "P3",
			}

			jsonData, _ := json.Marshal(payload)
			// Gọi đến Incident Service qua cổng 8005
			url := "http://incident-service:8005/api/v1/incidents/trigger"

			resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Printf("⚠️ Lỗi gửi sự kiện USB sang Incident Service: %v", err)
				return
			}
			defer resp.Body.Close()
		}(rec)
	}
	if len(activeHashes) > 0 {
		if _, err := database.USBCollection.UpdateMany(context.TODO(), bson.M{
			"asset_hwid":  asset.AssetHWID,
			"org_id":      int64(asset.OrgID),
			"event_type":  "CONNECTED",
			"device_hash": bson.M{"$nin": activeHashes},
		}, bson.M{"$set": bson.M{"event_type": "DISCONNECTED"}}); err != nil {
			return fmt.Errorf("lỗi cập nhật ngắt kết nối USB: %w", err)
		}
	}

	return nil
}
