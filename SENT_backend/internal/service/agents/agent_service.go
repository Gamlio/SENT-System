package agents

import (
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	dataAgents "sent_backend/internal/service/agents/data"
	"sent_backend/internal/websocket"
	"time"
)

type AgentPayload struct {
	Type     string      `json:"type"` // "DATA" | "HEARTBEAT"
	LogType  string      `json:"log_type"`
	HWID     string      `json:"hwid"`
	Hostname string      `json:"hostname"`
	Data     interface{} `json:"data"`
}

func ProcessAgentData(payload AgentPayload) {
	// 1. Cập nhật nhịp đập (Keep-alive)
	database.DB.Model(&models.Agent{}).Where("hw_id = ?", payload.HWID).
		Updates(map[string]interface{}{"last_seen": time.Now(), "hostname": payload.Hostname})

	if payload.Type == "HEARTBEAT" {
		return
	}

	var agent models.Agent
	if err := database.DB.Where("hw_id = ?", payload.HWID).First(&agent).Error; err != nil {
		return
	}

	// [CHỐT CHẶN BẢO MẬT]: Từ chối xử lý log nếu máy chưa được duyệt
	if agent.Status != "ACTIVE" {
		return
	}

	// 2. PHÂN LUỒNG XUỐNG CÁC MODULE CHUYÊN TRÁCH
	switch payload.LogType {
	case "software_baseline":
		dataAgents.HandleSoftwareBaseline(agent, payload.Data)
	case "software":
		dataAgents.ProcessSoftware(agent, payload.Data)
	case "usb":
		dataAgents.ProcessUSB(agent, payload.Data)
	case "port":
		dataAgents.ProcessPorts(agent, payload.Data)
	case "inventory":
		dataAgents.ProcessInventory(agent, payload.Data)
	case "firewall":
		dataAgents.ProcessFirewall(agent, payload.Data)
	case "antivirus":
		dataAgents.ProcessAntivirus(agent, payload.Data)
	}

	// Frontend AgentDetail.jsx sẽ nhận tin này và tự fetchDetail() lại
	websocket.GlobalHub.Broadcast(map[string]interface{}{
		"type": "AGENT_UPDATE",
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
