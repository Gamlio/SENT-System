package service

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"SENT_backend/pkg/websocket"
	dataassets "SENT_backend/service/assets/internal/service/data"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

type AssetPayload struct {
	Type     string      `json:"type" binding:"required,oneof=DATA HEARTBEAT ALERT OFFLINE"` // "DATA" | "HEARTBEAT" | "OFFLINE"
	LogType  string      `json:"log_type" binding:"required,max=50"`
	AssetID  string      `json:"asset_hwid" binding:"required,max=64"`
	Hostname string      `json:"hostname" binding:"required,min=1,max=255"`
	Data     interface{} `json:"data" binding:"required"`
}

func ProcessassetData(payload AssetPayload) error {

	// [TỐI ƯU HIỆU NĂNG] Giảm tải gõ búa (Hammering) cho Postgres bằng cách CHỈ cập nhật last_seen cho gói HEARTBEAT
	if payload.Type == "HEARTBEAT" {
		// Truy vấn nhanh OrgID để broadcast WebSocket chính xác
		var asset models.Asset
		err := database.DB.Select("org_id").Where("asset_hwid = ?", payload.AssetID).First(&asset).Error

		if err == nil {
			// Cập nhật last_seen cực nhanh không qua hooks
			database.DB.Model(&models.Asset{}).Where("asset_hwid = ?", payload.AssetID).UpdateColumn("last_seen", time.Now())

			// Phát tín hiệu Online cho Dashboard của đúng tổ chức (OrgID)
			websocket.GlobalHub.BroadcastToOrg(asset.OrgID, map[string]interface{}{
				"type":      "ASSET_HEARTBEAT",
				"hwid":      payload.AssetID,
				"last_seen": time.Now(),
			})
		}
		return err
	}

	// [LUỒNG TẮT MÁY NHANH]: Khi Agent gửi tín hiệu Logout/Shutdown
	if payload.Type == "OFFLINE" {
		var asset models.Asset
		database.DB.Select("org_id").Where("asset_hwid = ?", payload.AssetID).First(&asset)

		err := database.DB.Model(&models.Asset{}).Where("asset_hwid = ?", payload.AssetID).
			UpdateColumn("last_seen", time.Now().Add(-10*time.Minute)).Error // Ép Offline ngay
		if err == nil && asset.OrgID != 0 {
			websocket.GlobalHub.BroadcastToOrg(asset.OrgID, map[string]interface{}{"type": "ASSET_UPDATE", "hwid": payload.AssetID})
		}
		return err
	}

	var asset models.Asset
	if err := database.DB.Where("asset_hwid = ?", payload.AssetID).First(&asset).Error; err != nil {
		return fmt.Errorf("không tìm thấy thiết bị: %w", err)
	}

	// [CHỐT CHẶN BẢO MẬT]: Từ chối xử lý log nếu máy chưa được duyệt kích hoạt chính thức
	// Chấp nhận cả APPROVED vì máy vừa duyệt Baseline xong thường ở trạng thái này
	if asset.Status != "ACTIVE" && asset.Status != "APPROVED" {
		log.Printf("⚠️ [ProcessassetData] Từ chối xử lý dữ liệu cho HWID %s vì trạng thái: %s", payload.AssetID, asset.Status)
		return fmt.Errorf("thiết bị chưa được phê duyệt hoạt động (Status: %s)", asset.Status)
	}

	// 1. LUỒNG XỬ LÝ GÓI TIN GỘP BATCH TỪ AGENT (Định kỳ đẩy lên SOC)
	if strings.HasPrefix(payload.LogType, "batch_") {
		var batch map[string]json.RawMessage

		switch v := payload.Data.(type) {
		case json.RawMessage:
			if err := json.Unmarshal(v, &batch); err != nil {
				return fmt.Errorf("lỗi giải mã batch từ RawMessage: %w", err)
			}
		case []byte:
			if err := json.Unmarshal(v, &batch); err != nil {
				return fmt.Errorf("lỗi giải mã batch từ []byte: %w", err)
			}
		default:
			// Fallback trong trường hợp dữ liệu là map[string]interface{}
			bytes, _ := json.Marshal(payload.Data)
			if err := json.Unmarshal(bytes, &batch); err != nil {
				return fmt.Errorf("lỗi giải mã batch từ cấu trúc object: %w", err)
			}
		}

		// Duyệt qua từng module thành phần trong gói gộp batch dữ liệu gửi về từ Agent
		for modName, modData := range batch {
			// Phân luồng xuống các module phân tích dữ liệu chuyên trách hiện có
			switch modName {
			case "software":
				if err := dataassets.ProcessSoftware(asset, modData); err != nil {
					fmt.Printf("⚠️ Lỗi Software (%s): %v\n", asset.AssetHWID, err)
				}
			case "usb":
				if err := dataassets.ProcessUSB(asset, modData); err != nil {
					fmt.Printf("⚠️ Lỗi USB (%s): %v\n", asset.AssetHWID, err)
				}
			case "port":
				if err := dataassets.ProcessPorts(asset, modData); err != nil {
					fmt.Printf("⚠️ Lỗi Port (%s): %v\n", asset.AssetHWID, err)
				}
			case "inventory":
				if err := dataassets.ProcessInventory(asset, modData); err != nil {
					fmt.Printf("⚠️ Lỗi Inventory (%s): %v\n", asset.AssetHWID, err)
				}
			case "firewall":
				if err := dataassets.ProcessFirewall(asset, modData); err != nil {
					fmt.Printf("⚠️ Lỗi Firewall (%s): %v\n", asset.AssetHWID, err)
				}
			case "antivirus":
				if err := dataassets.ProcessAntivirus(asset, modData); err != nil {
					fmt.Printf("⚠️ Lỗi Antivirus (%s): %v\n", asset.AssetHWID, err)
				}
			case "data_transfer":
				if err := dataassets.ProcessDataTransfer(asset, modData); err != nil {
					fmt.Printf("⚠️ Lỗi DataTransfer (%s): %v\n", asset.AssetHWID, err)
				}
			case "patch":
				if err := dataassets.ProcessPatch(asset, modData); err != nil {
					fmt.Printf("⚠️ Lỗi Patch (%s): %v\n", asset.AssetHWID, err)
				}
			}
		}

		// Broadcast update; scoring recalculation will be triggered by Behavior Service only
		websocket.GlobalHub.BroadcastToOrg(asset.OrgID, map[string]interface{}{
			"type": "ASSET_UPDATE",
			"hwid": payload.AssetID,
		})

		return nil // Hoàn tất xử lý trọn vẹn gói tin gộp batch
	}

	// 2. LUỒNG PHÂN LUỒNG CHO CÁC GÓI TIN ĐƠN LẺ KHÔNG QUA BATCH (Log đẩy trực tiếp khẩn cấp)
	switch payload.LogType {
	case "software":
		if err := dataassets.ProcessSoftware(asset, payload.Data); err != nil {
			return err
		}
	case "usb":
		if err := dataassets.ProcessUSB(asset, payload.Data); err != nil {
			return err
		}
	case "port":
		if err := dataassets.ProcessPorts(asset, payload.Data); err != nil {
			return err
		}
	case "inventory":
		if err := dataassets.ProcessInventory(asset, payload.Data); err != nil {
			return err
		}
	case "firewall":
		if err := dataassets.ProcessFirewall(asset, payload.Data); err != nil {
			return err
		}
	case "antivirus":
		if err := dataassets.ProcessAntivirus(asset, payload.Data); err != nil {
			return err
		}
	case "data_transfer":
		if err := dataassets.ProcessDataTransfer(asset, payload.Data); err != nil {
			return err
		}
	case "patch":
		if err := dataassets.ProcessPatch(asset, payload.Data); err != nil {
			return err
		}
	}

	// Broadcast update; scoring recalculation should originate from Behavior Service
	websocket.GlobalHub.BroadcastToOrg(asset.OrgID, map[string]interface{}{
		"type": "ASSET_UPDATE",
		"hwid": payload.AssetID,
	})

	return nil
}
