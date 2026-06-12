package data

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

func ProcessETW(asset models.Asset, data interface{}) error {
	bytesData, err := GetBytesFromData(data)
	if err != nil {
		return fmt.Errorf("dữ liệu etw không hợp lệ: %w", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(bytesData, &payload); err != nil {
		return fmt.Errorf("lỗi giải mã JSON ETW: %w", err)
	}

	eventType := "etw"
	if t, ok := payload["type"].(string); ok && t != "" {
		eventType = t
	}

	if database.AssetETWCollection != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		rec := map[string]interface{}{
			"asset_hwid": asset.AssetHWID,
			"org_id":     int64(asset.OrgID),
			"event_type": eventType,
			"payload":    payload,
			"timestamp":  time.Now(),
		}
		_, _ = database.AssetETWCollection.InsertOne(ctx, rec)
	}

	// Simple heuristics -> Forward to Behavior Service for scoring
	// 1) Process starts
	if procs, ok := payload["processes"].([]interface{}); ok && len(procs) > 0 {
		for _, pi := range procs {
			if m, ok := pi.(map[string]interface{}); ok {
				pname := ""
				if v, ok := m["name"].(string); ok {
					pname = v
				}
				SendBehaviorLog("ProcessStart", pname, "Process started", fmt.Sprintf("Process %s started on asset", pname), "P3", asset)
			}
		}
	}

	// 2) Network connections (detect probable C2: remote public IPs)
	if nets, ok := payload["network"].([]interface{}); ok && len(nets) > 0 {
		for _, ni := range nets {
			if m, ok := ni.(map[string]interface{}); ok {
				rip, _ := m["remote_ip"].(string)
				if rip != "" {
					ip := net.ParseIP(rip)
					if ip != nil && !ip.IsPrivate() {
						SendBehaviorLog("C2Connection", rip, "Outbound to public IP", fmt.Sprintf("Connection to %s", rip), "P1", asset)
					}
				}
			}
		}
	}

	// 3) Registry changes are accepted and parsed by backend.
	if regs, ok := payload["registry"].([]interface{}); ok && len(regs) > 0 {
		SendBehaviorLog("RegistryChange", fmt.Sprintf("%d items", len(regs)), "Registry change observed", fmt.Sprintf("Registry change payload contained %d items", len(regs)), "P2", asset)
	}

	// 4) File IO surge -> possible ransomware
	if files, ok := payload["file_io"].([]interface{}); ok {
		if len(files) > 20 {
			SendBehaviorLog("RansomwareIO", "burst", "High file IO activity", fmt.Sprintf("%d file IO ops detected", len(files)), "P1", asset)
		}
	}

	return nil
}
