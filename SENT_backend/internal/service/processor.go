package service

import (
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/agent_data"
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
		agent_data.HandleSoftwareBaseline(agent, payload.Data)
	case "software":
		agent_data.ProcessSoftware(agent, payload.Data)
	case "usb":
		agent_data.ProcessUSB(agent, payload.Data)
	case "port":
		agent_data.ProcessPorts(agent, payload.Data)
	case "inventory":
		agent_data.ProcessInventory(agent, payload.Data)
	case "firewall":
		agent_data.ProcessFirewall(agent, payload.Data)
	case "antivirus":
		agent_data.ProcessAntivirus(agent, payload.Data)
	}
}
