package service

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"SENT_backend/pkg/websocket"
	dataassets "SENT_backend/service/assets/internal/service/data"
	"encoding/json"
	"fmt"
	"net/http"
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

	// [TỐI ƯU HIỆU NĂNG] Giảm tải gõ búa (Hammering) cho Postgres bằng cách CHỈ cập nhật last_seen cho gói HEARTBEAT
	if payload.Type == "HEARTBEAT" {
		return database.DB.Model(&models.Asset{}).Where("asset_hwid = ?", payload.AssetID).
			Update("last_seen", time.Now()).Error
	}

	var asset models.Asset
	if err := database.DB.Where("asset_hwid = ?", payload.AssetID).First(&asset).Error; err != nil {
		return fmt.Errorf("không tìm thấy thiết bị: %w", err)
	}

	// [CHỐT CHẶN BẢO MẬT]: Từ chối xử lý log nếu máy chưa được duyệt kích hoạt chính thức
	if asset.Status != "ACTIVE" {
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

			// [ĐỒNG BỘ BASELINE]: Kiểm tra xem máy có đang nằm trong 15 phút cấu hình máy sạch ban đầu hay không
			if asset.BaselineUntil != nil && time.Now().Before(*asset.BaselineUntil) {

				// Chỉ nạp Baseline tự động cho 3 danh mục chính sách tĩnh của Zero Trust Engine
				if modName == "software" || modName == "port" || modName == "usb" {
					var rawValues []map[string]interface{}
					var valuesToBaseline []map[string]interface{}

					// Giải mã an toàn mảng dữ liệu telemetry thô để trích xuất Value hoặc Hash
					if err := json.Unmarshal(modData, &rawValues); err == nil {
						for _, item := range rawValues {
							if modName == "software" && item["file_hash"] != nil {
								valuesToBaseline = append(valuesToBaseline, map[string]interface{}{
									"file_hash":     item["file_hash"],
									"software_name": item["software_name"],
									"publisher":     item["publisher"],
									"version":       item["version"],
								})
							} else if modName == "port" && item["port"] != nil {
								valuesToBaseline = append(valuesToBaseline, map[string]interface{}{
									"port":         item["port"],
									"process_name": item["process_name"],
								})
							} else if modName == "usb" && item["device_hash"] != nil {
								valuesToBaseline = append(valuesToBaseline, map[string]interface{}{
									"device_hash":   item["device_hash"],
									"device_name":   item["device_name"],
									"vid":           item["vid"],
									"pid":           item["pid"],
									"serial_number": item["serial_number"],
								})
							}
						}

						// Đẩy bất đồng bộ (Async Goroutine) sang Policy Service, tránh block luồng xử lý log chính
						if len(valuesToBaseline) > 0 {
							// Đồng bộ HOA tên Category (SOFTWARE_HASH, PORT, USB_DEVICE) tương thích với DB
							categoryType := strings.ToUpper(modName)
							if modName == "software" {
								categoryType = "SOFTWARE_HASH"
							} else if modName == "usb" {
								categoryType = "USB_DEVICE"
							}
							go dataassets.SendToPolicyBaseline(asset.OrgID, asset.AssetHWID, categoryType, valuesToBaseline)
						}
					}
				}
			}

			// Phân luồng xuống các module phân tích dữ liệu chuyên trách hiện có
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
			case "patch":
				if err := dataassets.ProcessPatch(asset, modData); err != nil {
					return err
				}
			}
		}

		// Kích hoạt tính lại điểm rủi ro qua Scoring Service
		go func(id string) {
			url := "http://scoring-service:8000/api/v1/scoring/recalculate/" + id
			_, _ = http.Post(url, "application/json", nil)
		}(payload.AssetID)

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

	// Phát tín hiệu WebSocket Real-time cập nhật trạng thái ra màn hình Dashboard của Web Admin
	websocket.GlobalHub.BroadcastToOrg(asset.OrgID, map[string]interface{}{
		"type": "ASSET_UPDATE",
		"hwid": payload.AssetID,
	})

	// Kích hoạt tính lại điểm rủi ro qua Scoring Service
	go func(id string) {
		url := "http://scoring-service:8000/api/v1/scoring/recalculate/" + id
		_, _ = http.Post(url, "application/json", nil)
	}(payload.AssetID)

	return nil
}
