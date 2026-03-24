package data

import (
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
)

func ProcessInventory(agent models.Agent, data interface{}) {
	invData, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	var inv models.AgentInventory
	result := database.DB.Where("agent_hw_id = ?", agent.HWID).First(&inv)

	inv.AgentHWID = agent.HWID
	inv.OSInfo = fmt.Sprintf("%v", invData["os_info"])
	inv.CPUModel = fmt.Sprintf("%v", invData["cpu_model"])

	if ramFloat, ok := invData["ram_total_gb"].(float64); ok {
		inv.RAMTotalGB = int(ramFloat)
	}

	if result.Error != nil {
		database.DB.Create(&inv)
	} else {
		database.DB.Save(&inv)
	}
}
