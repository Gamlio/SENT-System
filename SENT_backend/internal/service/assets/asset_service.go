package assets

import (
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	dataassets "sent_backend/internal/service/assets/data"
	"sent_backend/internal/websocket"
	"strings"
	"time"
)

type AssetPayload struct {
	Type     string      `json:"type" binding:"required,oneof=DATA HEARTBEAT ALERT"` // "DATA" | "HEARTBEAT"
	LogType  string      `json:"log_type" binding:"required,max=50"`
	AssetID  string      `json:"asset_hwid" binding:"required,max=64"`
	Hostname string      `json:"hostname" binding:"required,min=1,max=255"`
	Data     interface{} `json:"data" binding:"required"`
}

func ProcessassetData(payload AssetPayload) error {
	// [TỐI ƯU HIỆU NĂNG] Giảm tải (Hammering) cho Postgres bằng cách CHỈ cập nhật last_seen
	// khi nhận gói HEARTBEAT (chu kỳ 30s-1p/lần), bỏ qua việc update cho mỗi gói DATA gửi lên.
	if payload.Type == "HEARTBEAT" {
		return database.DB.Model(&models.Asset{}).Where("asset_hwid = ?", payload.AssetID).
			Update("last_seen", time.Now()).Error
	}

	var asset models.Asset
	if err := database.DB.Where("asset_hwid = ?", payload.AssetID).First(&asset).Error; err != nil {
		return fmt.Errorf("không tìm thấy thiết bị: %w", err)
	}

	// [CHỐT CHẶN BẢO MẬT]: Từ chối xử lý log nếu máy chưa được duyệt
	if asset.Status != "ACTIVE" {
		return fmt.Errorf("thiết bị chưa được phê duyệt hoạt động (Status: %s)", asset.Status)
	}

	// THÊM: Xử lý bóc tách cho các gói gộp batch_
	if strings.HasPrefix(payload.LogType, "batch_") {
		var batch map[string]json.RawMessage

		// Xử lý an toàn Data thô (RawMessage/Bytes) thành map các RawMessage con.
		// Việc này giúp truyền chính xác Mảng (Array) hoặc Đối tượng (Object) con vào hàm Process
		// mà không bị dính phím cha, đồng thời loại bỏ nguy cơ Double-Marshal làm vỡ chuỗi JSON.
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
			// Fallback trong trường hợp là map[string]interface{}
			bytes, _ := json.Marshal(payload.Data)
			if err := json.Unmarshal(bytes, &batch); err != nil {
				return fmt.Errorf("lỗi giải mã batch từ cấu trúc object: %w", err)
			}
		}

		// Phân phối dữ liệu vào các hàm chuyên biệt hiện có
		for modName, modData := range batch {
			switch modName {
			case "software":
				if err := dataassets.ProcessSoftware(asset, modData); err != nil {
					return err
				}
			case "usb":
				if err := dataassets.ProcessUSB(asset, modData); err != nil {
					return err
				}
			case "port":
				if err := dataassets.ProcessPorts(asset, modData); err != nil {
					return err
				}
			case "inventory":
				if err := dataassets.ProcessInventory(asset, modData); err != nil {
					return err
				}
			case "firewall":
				if err := dataassets.ProcessFirewall(asset, modData); err != nil {
					return err
				}
			case "antivirus":
				if err := dataassets.ProcessAntivirus(asset, modData); err != nil {
					return err
				}
			case "data_transfer":
				if err := dataassets.ProcessDataTransfer(asset, modData); err != nil {
					return err
				}
			}
		}
		return nil // Xử lý xong batch thì thoát hàm
	}

	// 2. PHÂN LUỒNG XUỐNG CÁC MODULE CHUYÊN TRÁCH
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
	}

	// Frontend assetDetail.jsx sẽ nhận tin này và tự fetchDetail() lại
	websocket.GlobalHub.BroadcastToOrg(asset.OrgID, map[string]interface{}{
		"type": "ASSET_UPDATE",
		"hwid": payload.AssetID,
	})

	return nil
}
