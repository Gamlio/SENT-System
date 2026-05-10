package data

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database" // Sửa từ pkg/models/database sang pkg/database
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	21:    "FTP (Cleartext credentials)",
	22:    "SSH (Remote Access)",
	23:    "Telnet (Insecure Remote Access)",
	25:    "SMTP (Email Relay/Spam)",
	53:    "DNS (Potential Tunneling)",
	111:   "RPCBind (Information Gathering)",
	135:   "RPC (Exploitation Target)",
	139:   "NetBIOS (Lateral Movement)",
	445:   "SMB (EternalBlue/WannaCry)",
	1433:  "MSSQL (Database Exposure)",
	27017: "MongoDB (Unauthenticated Access)",
	3306:  "MySQL (Database Exposure)",
	3389:  "RDP (Remote Desktop)",
	4444:  "Metasploit/Malware Backdoor",
	5432:  "PostgreSQL (Database Exposure)",
	5900:  "VNC (Remote Control)",
	6379:  "Redis (Unauthenticated Access)",
	8080:  "Proxy/Alternative HTTP",
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
			// GỌI API SANG INCIDENT SERVICE
			go func(p int, desc string, proc string) {
				event := map[string]interface{}{
					"asset":       asset,
					"alert_type":  "Unauthorized Port",
					"title":       fmt.Sprintf("[P2] Phát hiện cổng %d (%s)", p, desc),
					"description": fmt.Sprintf("Tiến trình '%s' đang mở cổng dịch vụ nhạy cảm %d.", proc, p),
					"priority":    "P2",
				}
				jsonData, _ := json.Marshal(event)
				// Gọi đến cổng 8005 của Incident Service
				http.Post("http://incident-service:8005/api/v1/incidents/trigger", "application/json", bytes.NewBuffer(jsonData))
			}(incomingPort.Port, riskDesc, incomingPort.ProcessName)
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
