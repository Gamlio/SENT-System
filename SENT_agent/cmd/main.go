package main

import (
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/host"

	"sent_agent/internal/analyzer" // <--- Import não phản xạ
	"sent_agent/internal/collector"
	"sent_agent/internal/config"
	"sent_agent/internal/transport"
)

func main() {
	if !config.Load() {
		config.PromptForCompanyCode()
	}

	hInfo, _ := host.Info()
	hwid := hInfo.HostID
	hostname, _ := os.Hostname()

	fmt.Printf("\n🛡️ AGENT V3.2 | COMPANY: %s | HOST: %s\n", config.Current.CompanyCode, hostname)

	client := transport.AgentClient

	// Các chính sách tải tạm về Agent (Thực tế nên lấy từ Server qua API)
	mockSoftwareBlacklist := []string{"bittorrent", "cheatengine", "coccoc"}
	mockUSBWhitelist := []string{"VID_045E"} // Ví dụ ID của chuột phím
	mockBlockedPorts := []int{3389, 21, 23}

	ticker := time.NewTicker(15 * time.Second)
	for range ticker.C {
		// 1. THU THẬP DỮ LIỆU
		telemetryData := collector.CollectTelemetry()
		softwareData := collector.CollectSoftware()
		usbData := collector.CollectUSB()

		// 2. [AGENTIC WORKFLOW] NÃO PHẢN XẠ TỰ PHÂN TÍCH & TÍNH ĐIỂM
		analyzer.LocalRiskScore = 0 // Reset điểm mỗi chu kỳ

		analyzer.EvaluateSoftware(softwareData, mockSoftwareBlacklist)
		analyzer.EvaluateUSB(usbData, mockUSBWhitelist)
		analyzer.EvaluateOpenPorts(telemetryData, mockBlockedPorts)
		analyzer.EvaluateDualHoming(hwid, hostname)
		analyzer.EvaluateAntivirus(hwid, hostname) // Uncomment if you have the function implemented

		// Nếu máy quá nguy hiểm, tăng tần suất gửi lên Server (Thích nghi - Adaptive)
		forceSend := false
		if analyzer.LocalRiskScore >= 50 {
			forceSend = true
			fmt.Println("🔥 Mức độ rủi ro cao, ÉP GỬI dữ liệu khẩn cấp về SOC!")
		}

		// 3. GỬI DỮ LIỆU VỀ SERVER
		client.SendPayload(hwid, hostname, "telemetry", telemetryData, forceSend)
		client.SendPayload(hwid, hostname, "software", softwareData, forceSend)
		client.SendPayload(hwid, hostname, "usb", usbData, forceSend)
	}
}
