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

	ticker := time.NewTicker(45 * time.Second)
	loopCount := 0 // Thêm biến đếm

	for range ticker.C {
		loopCount++

		// 1. THU THẬP DỮ LIỆU (15 giây / lần)
		telemetryData := collector.CollectTelemetry()
		softwareData := collector.CollectSoftware()
		usbData := collector.CollectUSB()

		// 2. [AGENTIC WORKFLOW] NÃO PHẢN XẠ TỰ PHÂN TÍCH
		analyzer.LocalRiskScore = 0

		analyzer.EvaluateSoftware(softwareData, mockSoftwareBlacklist)
		analyzer.EvaluateUSB(usbData, mockUSBWhitelist)
		analyzer.EvaluateOpenPorts(telemetryData, mockBlockedPorts)

		// Các task nặng chạy ít hơn (Ví dụ mỗi 1 phút / 4 chu kỳ)
		if loopCount%4 == 0 {
			// BƯỚC 1: THU THẬP THÔNG TIN
			fwData := collector.CollectFirewall()
			avData := collector.CollectAntivirus()

			// BƯỚC 2: PHÂN TÍCH VÀ CẢNH BÁO
			analyzer.EvaluateFirewall(hwid, hostname, fwData)
			analyzer.EvaluateAntivirus(hwid, hostname, avData)
		}

		// Gửi dữ liệu khẩn cấp hoặc định kỳ...
		forceSend := false
		if analyzer.LocalRiskScore >= 50 {
			forceSend = true
			fmt.Println("⚠️ Điểm rủi ro cao! Bắt buộc gửi Log về SOC ngay lập tức.")
		}

		client.SendPayload(hwid, hostname, "telemetry", telemetryData, forceSend)
		client.SendPayload(hwid, hostname, "software", softwareData, forceSend)
		client.SendPayload(hwid, hostname, "usb", usbData, forceSend)
	}
}
