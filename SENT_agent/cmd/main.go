package main

import (
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/host"

	"sent_agent/internal/analyzer"
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

	fmt.Printf("\n AGENT V3.2.9 (Zero Trust) | COMPANY: %s | HOST: %s\n", config.Current.CompanyCode, hostname)

	client := transport.AgentClient

	mockSoftwareBlacklist := []string{"bittorrent", "cheatengine", "coccoc"}
	mockUSBWhitelist := []string{"VID_045E"}
	mockBlockedPorts := []int{3389, 21, 23}

	ticker := time.NewTicker(45 * time.Second) // Có thể đổi thành 15s để test nhanh hơn
	loopCount := 0

	// [MỚI] Biến kiểm soát trạng thái Zero Trust
	isPending := false
	alertShown := false

	for range ticker.C {
		loopCount++

		// --- CHẾ ĐỘ NGỦ ĐÔNG CHỜ PHÊ DUYỆT ---
		if isPending {
			fmt.Println("⏳ Máy đang ở trạng thái chờ duyệt. Gửi ping kiểm tra...")
			// Chỉ gửi 1 gói telemetry nhỏ lên để check xem admin duyệt chưa
			telemetryData := collector.CollectTelemetry()
			serverState := client.SendPayload(hwid, hostname, "telemetry", telemetryData, true)

			if serverState == "PENDING" {
				if !alertShown {
					analyzer.ShowSystemAlert("ĐANG CHỜ PHÊ DUYỆT", "Hệ thống bảo mật SENT đã ghi nhận thiết bị. Vui lòng chờ Quản trị viên phê duyệt.")
					alertShown = true
				}
				continue // Bỏ qua toàn bộ logic bên dưới, quay lại vòng lặp chờ
			} else if serverState == "REJECTED" {
				fmt.Println("Admin đã từ chối máy trạm này.")
				continue
			} else if serverState == "ACTIVE" || serverState == "OK" {
				fmt.Println("ADMIN ĐÃ DUYỆT! Hệ thống bắt đầu bảo vệ.")
				isPending = false
				analyzer.ShowSystemAlert("ĐÃ PHÊ DUYỆT", "Thiết bị của bạn đã được phép tham gia mạng an toàn.")
			}
		}

		// --- CHẾ ĐỘ HOẠT ĐỘNG BÌNH THƯỜNG (ACTIVE) ---
		// 1. THU THẬP DỮ LIỆU
		telemetryData := collector.CollectTelemetry()
		softwareData := collector.CollectSoftware()
		usbData := collector.CollectUSB()

		// 2. [AGENTIC WORKFLOW] NÃO PHẢN XẠ TỰ PHÂN TÍCH
		analyzer.LocalRiskScore = 0
		analyzer.EvaluateSoftware(softwareData, mockSoftwareBlacklist)
		analyzer.EvaluateUSB(usbData, mockUSBWhitelist)
		analyzer.EvaluateOpenPorts(telemetryData, mockBlockedPorts)

		// 3. GỬI LÊN SERVER (Và kiểm tra phản hồi)
		state1 := client.SendPayload(hwid, hostname, "telemetry", telemetryData, false)
		client.SendPayload(hwid, hostname, "software", softwareData, false)
		client.SendPayload(hwid, hostname, "usb", usbData, false)

		// Nếu server bất ngờ đá ra ngoài (VD: Admin bấm "Cách ly")
		if state1 == "PENDING" || state1 == "ISOLATED" {
			isPending = true
			alertShown = false
			continue
		}

		// Các task nặng chạy ít hơn (Ví dụ mỗi 1 phút / 4 chu kỳ)
		if loopCount%4 == 0 {
			fwData := collector.CollectFirewall()
			avData := collector.CollectAntivirus()

			analyzer.CheckAntivirusStatus(hwid, hostname)
			analyzer.CheckFirewallStatus(hwid, hostname)

			client.SendPayload(hwid, hostname, "firewall", fwData, false)
			client.SendPayload(hwid, hostname, "antivirus", avData, false)

			// Gửi dữ liệu tĩnh
			invData := collector.CollectInventory()
			client.SendPayload(hwid, hostname, "inventory", invData, false)
		}
	}
}
