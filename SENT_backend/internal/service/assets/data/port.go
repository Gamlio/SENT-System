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
	bytes, _ := json.Marshal(data)
	var payload assetTelemetryRecord
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return
	}
	if database.OpenPortCollection == nil {
		return
	}

	incSvc := &incidents.IncidentService{}
	incomingPortsMap := make(map[int]bool)

	for _, incomingPort := range payload.OpenPorts {
		incomingPortsMap[incomingPort.Port] = true
		filter := bson.M{"asset_hwid": asset.HWID, "port": incomingPort.Port}
		update := bson.M{"$set": bson.M{
			"asset_hwid":   asset.HWID,
			"port":         incomingPort.Port,
			"process_name": incomingPort.ProcessName,
			"status":       "OPEN",
			"updated_at":   time.Now(),
		}}

		_, _ = database.OpenPortCollection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))

		if incomingPort.Port == 3389 || incomingPort.Port == 22 || incomingPort.Port == 4444 {
			incSvc.TriggerSecurityEvent(asset, "Unauthorized Port",
				fmt.Sprintf("[P2] Mở cổng quản trị (%d) trái phép", incomingPort.Port),
				fmt.Sprintf("Tiến trình '%s' đang mở cổng %d.", incomingPort.ProcessName, incomingPort.Port), "P2")
		}
	}

	cursor, err := database.OpenPortCollection.Find(context.TODO(), bson.M{"asset_hwid": asset.HWID, "status": "OPEN"})
	if err != nil {
		return
	}
	var activeDBPorts []models.OpenPort
	cursor.All(context.TODO(), &activeDBPorts)

	for _, dbPort := range activeDBPorts {
		if !incomingPortsMap[dbPort.Port] {
			_, _ = database.OpenPortCollection.UpdateOne(context.TODO(), bson.M{"_id": dbPort.ID}, bson.M{"$set": bson.M{"status": "CLOSED"}})
		}
	}
}
