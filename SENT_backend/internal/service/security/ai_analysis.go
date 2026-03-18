package security

import (
	"fmt"
	"log"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"strings"
)

// AnalyzeBehaviorAI: Phân tích hành vi bất thường
func AnalyzeBehaviorAI(hwid string, dataType string, data interface{}) {
	var agent models.Agent
	if err := database.DB.Where("hw_id = ?", hwid).First(&agent).Error; err != nil {
		return
	}

	log.Printf("🤖 AI đang phân tích dữ liệu [%s] của máy %s...", dataType, agent.Hostname)

	switch dataType {
	case "telemetry":
		analyzeNetworkAI(agent, data)
	case "software":
		analyzeSoftwareAI(agent, data)
	}
}

func analyzeNetworkAI(agent models.Agent, data interface{}) {
	payload, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	ports, ok := payload["open_ports"].([]interface{})
	if !ok {
		return
	}

	for _, p := range ports {
		portMap, _ := p.(map[string]interface{})
		portNum := int(portMap["port"].(float64))

		// RDP Port -> Tạo Alert -> Tự động sinh Incident kèm Playbook xử lý Port lạ
		if portNum == 3389 {
			TriggerSecurityEvent(agent, "Unauthorized Port", "Rủi ro RDP (3389)",
				"AI phát hiện cổng Remote Desktop đang mở public. Nguy cơ tấn công cao.", "Critical")
		}
		if portNum == 23 {
			TriggerSecurityEvent(agent, "Unauthorized Port", "Giao thức Telnet (23)",
				"Cổng Telnet không an toàn đang mở.", "High")
		}
	}
}

func analyzeSoftwareAI(agent models.Agent, data interface{}) {
	softwareList, ok := data.([]interface{})
	if !ok {
		return
	}

	for _, item := range softwareList {
		swMap, _ := item.(map[string]interface{})
		name := strings.ToLower(fmt.Sprintf("%v", swMap["software_name"]))

		// Ví dụ logic AI đơn giản (Sau này thay bằng Model thật)
		if strings.Contains(name, "miner") || strings.Contains(name, "hack") {
			TriggerSecurityEvent(agent, "AI_MALWARE_SUSPICION", "Nghi ngờ phần mềm độc hại",
				fmt.Sprintf("AI phát hiện phần mềm có tên khả nghi: %s", name), "High")
		}
	}
}
