package data // Đổi sang package data

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/incidents"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type assetTelemetryRecord struct {
	IPAddress   string            `json:"ip_address"`
	FirewallOff bool              `json:"firewall_off"`
	OpenPorts   []models.OpenPort `json:"open_ports"`
}

func ProcessPorts(asset models.Asset, data interface{}) error {
	bytes, err := GetBytesFromData(data)
	if err != nil {
		return fmt.Errorf("dữ liệu port không hợp lệ: %w", err)
	}
	log.Printf("[DEBUG] HWID: %s | Bytes: %s", asset.AssetHWID, string(bytes))
	var payload assetTelemetryRecord
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return fmt.Errorf("lỗi giải mã JSON port: %w", err)
	}
	if database.OpenPortCollection == nil {
		return fmt.Errorf("database OpenPortCollection chưa sẵn sàng")
	}

	incSvc := &incidents.IncidentService{}
	var activePorts []int

	for _, incomingPort := range payload.OpenPorts {
		activePorts = append(activePorts, incomingPort.Port)
		filter := bson.M{"asset_hwid": asset.AssetHWID, "org_id": int64(asset.OrgID), "port": incomingPort.Port}
		update := bson.M{"$set": bson.M{
			"asset_hwid":   asset.AssetHWID,
			"org_id":       int64(asset.OrgID),
			"port":         incomingPort.Port,
			"process_name": incomingPort.ProcessName,
			"status":       "OPEN",
			"updated_at":   time.Now(),
		}}

		if _, err := database.OpenPortCollection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true)); err != nil {
			return fmt.Errorf("lỗi cập nhật cổng mở trên MongoDB: %w", err)
		}

		if incomingPort.Port == 3389 || incomingPort.Port == 22 || incomingPort.Port == 4444 {
			incSvc.TriggerSecurityEvent(context.TODO(), asset, "Unauthorized Port",
				fmt.Sprintf("[P2] Mở cổng quản trị (%d) trái phép", incomingPort.Port),
				fmt.Sprintf("Tiến trình '%s' đang mở cổng %d.", incomingPort.ProcessName, incomingPort.Port), "P2")
		}
	}

	// TỐI ƯU ĐÓNG TRẠNG THÁI:
	if len(activePorts) > 0 {
		if _, err := database.OpenPortCollection.UpdateMany(context.TODO(), bson.M{
			"asset_hwid": asset.AssetHWID,
			"org_id":     int64(asset.OrgID),
			"status":     "OPEN",
			"port":       bson.M{"$nin": activePorts},
		}, bson.M{"$set": bson.M{"status": "CLOSED"}}); err != nil {
			return fmt.Errorf("lỗi đóng trạng thái các cổng cũ: %w", err)
		}
	}
	return nil
}
