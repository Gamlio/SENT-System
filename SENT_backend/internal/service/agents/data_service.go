package agents

import (
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"time"
)

type AgentDataService struct{}

// GetAgentStats: Tính toán số liệu tổng quan cho Dashboard
func (s *AgentDataService) GetAgentStats() (map[string]int64, error) {
	var total, online, alerts, regions int64

	database.DB.Model(&models.Agent{}).Count(&total)

	threshold := time.Now().Add(-2 * time.Minute)
	database.DB.Model(&models.Agent{}).Where("last_seen >= ?", threshold).Count(&online)

	database.DB.Model(&models.SecurityAlert{}).Where("is_resolved = ?", false).Count(&alerts)
	database.DB.Model(&models.Region{}).Count(&regions)

	return map[string]int64{
		"total": total, "online": online, "alerts": alerts, "regions": regions,
	}, nil
}

// GetAgentList: Lấy danh sách máy kèm logic Online/Offline ảo
func (s *AgentDataService) GetAgentList(orgID uint) []models.Agent {
	var agents []models.Agent
	database.DB.Preload("Manager").Preload("Inventory").
		Where("org_id = ? AND status != ?", orgID, "RETIRED").
		Order("last_seen desc").Find(&agents)

	threshold := time.Now().Add(-2 * time.Minute)
	for i := range agents {
		if agents[i].Status == "ACTIVE" {
			if agents[i].LastSeen.After(threshold) {
				agents[i].Status = "online"
			} else {
				agents[i].Status = "offline"
			}
		}
	}
	return agents
}

// GetAgentDetail: Truy vấn sâu 1 máy trạm
func (s *AgentDataService) GetAgentDetail(hwid string) (models.Agent, error) {
	var agent models.Agent
	err := database.DB.Preload("Inventory").Preload("Software").
		Preload("Alerts").Preload("USBLogs").Preload("OpenPorts").
		Preload("Manager").Where("hw_id = ?", hwid).First(&agent).Error
	return agent, err
}

// GetAgentLogs: Lấy lịch sử cảnh báo của 1 máy
func (s *AgentDataService) GetAgentLogs(hwid string) []models.SecurityAlert {
	var alerts []models.SecurityAlert
	database.DB.Where("hw_id = ?", hwid).Order("created_at desc").Find(&alerts)
	return alerts
}
