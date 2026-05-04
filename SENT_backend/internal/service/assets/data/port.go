package data // Đổi sang package data

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

type assetTelemetryRecord struct {
	IPAddress   string            `json:"ip_address"`
	FirewallOff bool              `json:"firewall_off"`
	OpenPorts   []models.OpenPort `json:"open_ports"`
}

func ProcessPorts(asset models.Asset, data interface{}) {
	bytes, err := GetBytesFromData(data)
	if err != nil {
		return
	}

	var payload assetTelemetryRecord
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return
	}
	if database.OpenPortCollection == nil {
		return
	}

	incSvc := &incidents.IncidentService{}
	var activePorts []int

	for _, incomingPort := range payload.OpenPorts {
		activePorts = append(activePorts, incomingPort.Port)
		filter := bson.M{"asset_hwid": asset.AssetHWID, "org_id": asset.OrgID, "port": incomingPort.Port}
		update := bson.M{"$set": bson.M{
			"asset_hwid":   asset.AssetHWID,
			"org_id":       asset.OrgID,
			"port":         incomingPort.Port,
			"process_name": incomingPort.ProcessName,
			"status":       "OPEN",
			"updated_at":   time.Now(),
		}}

		_, _ = database.OpenPortCollection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))

		if incomingPort.Port == 3389 || incomingPort.Port == 22 || incomingPort.Port == 4444 {
			incSvc.TriggerSecurityEvent(context.TODO(), asset, "Unauthorized Port",
				fmt.Sprintf("[P2] Mở cổng quản trị (%d) trái phép", incomingPort.Port),
				fmt.Sprintf("Tiến trình '%s' đang mở cổng %d.", incomingPort.ProcessName, incomingPort.Port), "P2")
		}
	}

	// TỐI ƯU ĐÓNG TRẠNG THÁI:
	if len(activePorts) > 0 {
		_, _ = database.OpenPortCollection.UpdateMany(context.TODO(), bson.M{
			"asset_hwid": asset.AssetHWID,
			"org_id":     asset.OrgID,
			"status":     "OPEN",
			"port":       bson.M{"$nin": activePorts},
		}, bson.M{"$set": bson.M{"status": "CLOSED"}})
	}
}
