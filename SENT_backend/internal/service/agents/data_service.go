package agents

import (
	"context"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type AgentDataService struct{}

// GetAgentStats: Tính toán số liệu tổng quan cho Dashboard
func (s *AgentDataService) GetAgentStats() (map[string]int64, error) {
	var total, online, regions int64
	var alerts int64

	database.DB.Model(&models.Agent{}).Count(&total)

	threshold := time.Now().Add(-2 * time.Minute)
	database.DB.Model(&models.Agent{}).Where("last_seen >= ?", threshold).Count(&online)

	if database.SecurityAlertCollection != nil {
		c, err := database.SecurityAlertCollection.CountDocuments(context.TODO(), bson.M{"is_resolved": false})
		if err != nil {
			return nil, err
		}
		alerts = c
	}

	database.DB.Model(&models.Region{}).Count(&regions)

	return map[string]int64{
		"total": total, "online": online, "alerts": alerts, "regions": regions,
	}, nil
}

func (s *AgentDataService) attachAgentTelemetry(agent *models.Agent) {
	ctx := context.TODO()

	if database.AgentInventoryCollection != nil {
		var inv models.AgentInventory
		err := database.AgentInventoryCollection.FindOne(ctx, bson.M{"agent_hwid": agent.HWID}).Decode(&inv)
		if err == nil {
			agent.Inventory = inv
		}
	}

	if database.SoftwareCollection != nil {
		cursor, err := database.SoftwareCollection.Find(ctx, bson.M{"agent_hwid": agent.HWID})
		if err == nil {
			var items []models.SoftwareItem
			cursor.All(ctx, &items)
			agent.Software = items
		}
	}

	if database.SecurityAlertCollection != nil {
		cursor, err := database.SecurityAlertCollection.Find(ctx, bson.M{"hw_id": agent.HWID})
		if err == nil {
			var alerts []models.SecurityAlert
			cursor.All(ctx, &alerts)
			agent.Alerts = alerts
		}
	}

	if database.OpenPortCollection != nil {
		cursor, err := database.OpenPortCollection.Find(ctx, bson.M{"agent_hwid": agent.HWID})
		if err == nil {
			var ports []models.OpenPort
			cursor.All(ctx, &ports)
			agent.OpenPorts = ports
		}
	}

	if database.USBCollection != nil {
		cursor, err := database.USBCollection.Find(ctx, bson.M{"agent_hwid": agent.HWID})
		if err == nil {
			var usb []models.USBLog
			cursor.All(ctx, &usb)
			agent.USBLogs = usb
		}
	}

	if database.AgentIOActivityCollection != nil {
		cursor, err := database.AgentIOActivityCollection.Find(ctx, bson.M{"agent_hwid": agent.HWID})
		if err == nil {
			var ios []models.AgentIOActivity
			cursor.All(ctx, &ios)
			agent.IOActivities = ios
		}
	}
}

// GetAgentList: Lấy danh sách máy kèm logic Online/Offline ảo
func (s *AgentDataService) GetAgentList(orgID uint) []models.Agent {
	var agents []models.Agent
	database.DB.Preload("Manager").
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
		s.attachAgentTelemetry(&agents[i])
	}
	return agents
}

// GetAgentDetail: Truy vấn sâu 1 máy trạm
func (s *AgentDataService) GetAgentDetail(hwid string) (models.Agent, error) {
	var agent models.Agent
	err := database.DB.Preload("Manager").Where("hw_id = ?", hwid).First(&agent).Error
	if err != nil {
		return agent, err
	}

	s.attachAgentTelemetry(&agent)
	return agent, nil
}

// GetAgentLogs: Lấy lịch sử cảnh báo của 1 máy
func (s *AgentDataService) GetAgentLogs(hwid string) []models.SecurityAlert {
	var alerts []models.SecurityAlert
	if database.SecurityAlertCollection != nil {
		cursor, err := database.SecurityAlertCollection.Find(context.TODO(), bson.M{"hw_id": hwid})
		if err == nil {
			cursor.All(context.TODO(), &alerts)
		}
	}
	return alerts
}
