package assets

import (
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	dataassets "sent_backend/internal/service/assets/data"
	"sent_backend/internal/websocket"
	"time"
)

type AssetPayload struct {
	Type     string      `json:"type" binding:"required,oneof=DATA HEARTBEAT"` // "DATA" | "HEARTBEAT"
	LogType  string      `json:"log_type" binding:"required,max=50,alphanum"`
	HWID     string      `json:"hwid" binding:"required,max=64,alphanum"`
	Hostname string      `json:"hostname" binding:"required,min=1,max=255"`
	Data     interface{} `json:"data" binding:"required"`
}

func ProcessassetData(payload AssetPayload) {
	// 1. Cập nhật nhịp đập (Keep-alive)
	database.DB.Model(&models.Asset{}).Where("hw_id = ?", payload.HWID).
		Updates(map[string]interface{}{"last_seen": time.Now(), "hostname": payload.Hostname})

	if payload.Type == "HEARTBEAT" {
		return
	}

	var asset models.Asset
	if err := database.DB.Where("hw_id = ?", payload.HWID).First(&asset).Error; err != nil {
		return
	}

	// [CHỐT CHẶN BẢO MẬT]: Từ chối xử lý log nếu máy chưa được duyệt
	if asset.Status != "ACTIVE" {
		return
	}

	// 2. PHÂN LUỒNG XUỐNG CÁC MODULE CHUYÊN TRÁCH
	switch payload.LogType {
	case "software_baseline":
		dataassets.HandleSoftwareBaseline(asset, payload.Data)
	case "software":
		dataassets.ProcessSoftware(asset, payload.Data)
	case "usb":
		dataassets.ProcessUSB(asset, payload.Data)
	case "port":
		dataassets.ProcessPorts(asset, payload.Data)
	case "inventory":
		dataassets.ProcessInventory(asset, payload.Data)
	case "firewall":
		dataassets.ProcessFirewall(asset, payload.Data)
	case "antivirus":
		dataassets.ProcessAntivirus(asset, payload.Data)
	case "data_transfer":
		dataassets.ProcessDataTransfer(asset, payload.Data)
	}

	// Frontend assetDetail.jsx sẽ nhận tin này và tự fetchDetail() lại
	websocket.GlobalHub.Broadcast(map[string]interface{}{
		"type": "ASSET_UPDATE",
		"hwid": payload.HWID,
	})

	// Nếu là Baseline xong, báo tin riêng
	if payload.LogType == "software_baseline" {
		websocket.GlobalHub.Broadcast(map[string]interface{}{
			"type": "BASELINE_COMPLETED",
			"hwid": payload.HWID,
		})
	}
}
