package security

import (
	"fmt"
	"log"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"strings"
)

// AnalyzeBehaviorAI: Phân tích hành vi bất thường dựa trên dữ liệu gửi lên
func AnalyzeBehaviorAI(hwid string, dataType string, data interface{}) {
	// 1. Lấy thông tin Agent để phục vụ tạo cảnh báo
	var agent models.Agent
	if err := database.DB.Where("hw_id = ?", hwid).First(&agent).Error; err != nil {
		return
	}

	log.Printf("🤖 AI đang phân tích dữ liệu [%s] của máy %s...", dataType, agent.Hostname)

	// 2. Phân tích theo từng loại dữ liệu
	switch dataType {
	case "telemetry":
		analyzeNetworkAI(agent, data)
	case "software":
		analyzeSoftwareAI(agent, data)
	}
}

// --- CÁC HÀM PHÂN TÍCH CHI TIẾT ---

// AI phát hiện cổng mạng nguy hiểm
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
		portNum := int(portMap["port"].(float64)) // JSON number luôn là float64

		// Luật AI: Cảnh báo nếu mở cổng RDP (3389) hoặc Telnet (23)
		if portNum == 3389 {
			CreateAlert(agent, "AI_NETWORK_RISK", "Rủi ro truy cập từ xa",
				"AI phát hiện cổng Remote Desktop (3389) đang mở. Nguy cơ bị tấn công Brute-force cao.", "Critical")
		}
		if portNum == 23 {
			CreateAlert(agent, "AI_NETWORK_RISK", "Giao thức không an toàn",
				"AI phát hiện cổng Telnet (23) đang mở. Dữ liệu truyền đi không được mã hóa.", "High")
		}
	}
}

// AI phát hiện phần mềm đào coin hoặc hack game
func analyzeSoftwareAI(agent models.Agent, data interface{}) {
	softwareList, ok := data.([]interface{})
	if !ok {
		return
	}

	for _, item := range softwareList {
		swMap, _ := item.(map[string]interface{})
		name := strings.ToLower(fmt.Sprintf("%v", swMap["software_name"]))

		// Luật AI: Quét từ khóa nhạy cảm
		if strings.Contains(name, "miner") || strings.Contains(name, "xmrig") {
			CreateAlert(agent, "AI_MALWARE_DETECT", "Nghi vấn đào tiền ảo",
				"AI phát hiện phần mềm có dấu hiệu đào coin: "+name, "Critical")
		}

		if strings.Contains(name, "cheat") || strings.Contains(name, "hack") {
			CreateAlert(agent, "AI_POLICY_VIOLATION", "Phần mềm gian lận",
				"AI phát hiện công cụ gian lận/hack: "+name, "Medium")
		}
	}
}
