package agent_data

import (
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/security" // Import security
)

func ProcessSoftware(agent models.Agent, data interface{}) {
	softwareData, ok := data.([]interface{})
	if !ok {
		return
	}

	// Xóa dữ liệu cũ để cập nhật mới
	database.DB.Where("agent_hw_id = ?", agent.HWID).Delete(&models.SoftwareItem{})

	for _, item := range softwareData {
		swMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		swName := fmt.Sprintf("%v", swMap["software_name"])
		swVersion := fmt.Sprintf("%v", swMap["version"])

		dbItem := models.SoftwareItem{
			AgentHWID:    agent.HWID,
			SoftwareName: swName,
			Version:      swVersion,
		}
		database.DB.Create(&dbItem)
	}

	// QUAN TRỌNG: Gọi hàm kiểm tra chính sách sau khi lưu
	go security.CheckSoftwareCompliance(agent, data)
}
