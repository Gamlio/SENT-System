package data

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type assetTelemetryRecord struct {
	IPAddress   string            `json:"ip_address"`
	FirewallOff bool              `json:"firewall_off"`
	OpenPorts   []models.OpenPort `json:"open_ports"`
}

var dangerousPorts = map[int]string{
	21:    "FTP",
	22:    "SSH ",
	23:    "Telnet",
	25:    "SMTP",
	53:    "DNS",
	111:   "RPCBind",
	135:   "RPC",
	139:   "NetBIOS ",
	445:   "SMB ",
	1433:  "MSSQL ",
	27017: "MongoDB",
	3306:  "MySQL",
	3389:  "RDP",
	4444:  "Metasploit",
	5432:  "PostgreSQL",
	5900:  "VNC",
	6379:  "Redis",
	8080:  "Proxy",
}

func ProcessPorts(asset models.Asset, data interface{}) error {
	bytesData, err := GetBytesFromData(data)
	if err != nil {
		return fmt.Errorf("dữ liệu port không hợp lệ: %w", err)
	}

	var payload assetTelemetryRecord
	if err := json.Unmarshal(bytesData, &payload); err != nil {
		return fmt.Errorf("lỗi giải mã JSON port: %w", err)
	}

	// NHÁNH RIÊNG CHO POLICY BASELINE: Không ảnh hưởng đến logic bên dưới
	if asset.BaselineUntil != nil && time.Now().Before(*asset.BaselineUntil) {
		var entries []map[string]interface{}
		for _, p := range payload.OpenPorts {
			entries = append(entries, map[string]interface{}{
				"port":         p.Port,
				"process_name": p.ProcessName,
			})
		}
		SendToPolicyBaseline(asset.OrgID, asset.AssetHWID, "PORT", entries)
	}

	var activePorts []int
	for _, incomingPort := range payload.OpenPorts {
		activePorts = append(activePorts, incomingPort.Port)

		// 1. Cập nhật trạng thái cổng vào MongoDB
		filter := bson.M{"asset_hwid": asset.AssetHWID, "org_id": int64(asset.OrgID), "port": incomingPort.Port}
		update := bson.M{"$set": bson.M{
			"asset_hwid":   asset.AssetHWID,
			"org_id":       int64(asset.OrgID),
			"port":         incomingPort.Port,
			"process_name": incomingPort.ProcessName,
			"status":       "OPEN",
			"updated_at":   time.Now(),
		}}

		if database.OpenPortCollection != nil {
			database.OpenPortCollection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
		}

		// 2. Kiểm tra nếu cổng nằm trong danh sách nguy hiểm
		if riskDesc, isDangerous := dangerousPorts[incomingPort.Port]; isDangerous {
			// GỌI API SANG BEHAVIOR SERVICE
			SendBehaviorLog(
				"Unauthorized Port",
				fmt.Sprintf("%d", incomingPort.Port),
				fmt.Sprintf("[P2] Phát hiện cổng %d (%s)", incomingPort.Port, riskDesc),
				fmt.Sprintf("Tiến trình '%s' đang mở cổng dịch vụ nhạy cảm %d.", incomingPort.ProcessName, incomingPort.Port),
				"P2",
				asset,
			)
		}
	}

	// 3. Đóng các cổng không còn xuất hiện trong lần báo cáo này
	if len(activePorts) > 0 && database.OpenPortCollection != nil {
		database.OpenPortCollection.UpdateMany(context.TODO(), bson.M{
			"asset_hwid": asset.AssetHWID,
			"org_id":     int64(asset.OrgID),
			"status":     "OPEN",
			"port":       bson.M{"$nin": activePorts},
		}, bson.M{"$set": bson.M{"status": "CLOSED"}})
	}

	return nil
}
