package main

import (
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/host"

	// [SỬA LỖI TẠI ĐÂY] Tuyệt đối không dùng ../
	"sent_agent/internal/collector"
	"sent_agent/internal/config"
	"sent_agent/internal/transport"
)

func main() {
	// 1. Load Cấu hình
	if !config.Load() {
		config.PromptForCompanyCode()
	}

	// 2. Lấy thông tin máy
	hInfo, _ := host.Info()
	hwid := hInfo.HostID
	hostname, _ := os.Hostname()

	fmt.Printf("\n🛡️  AGENT STARTED | COMPANY: %s | HOST: %s\n", config.Current.CompanyCode, hostname)

	// 3. Gửi dữ liệu lần đầu (Force = true)
	client := transport.AgentClient

	// Gọi các hàm thu thập từ package collector (đã viết hoa chữ cái đầu)
	client.SendPayload(hwid, hostname, "inventory", collector.CollectInventory(), true)
	client.SendPayload(hwid, hostname, "software", collector.CollectSoftware(), true)
	client.SendPayload(hwid, hostname, "telemetry", collector.CollectTelemetry(), true)
	client.SendPayload(hwid, hostname, "usb", collector.CollectUSB(), true)

	// 4. Vòng lặp định kỳ (Ticker)
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		// Chỉ gửi nếu có sự thay đổi (Force = false)
		client.SendPayload(hwid, hostname, "telemetry", collector.CollectTelemetry(), false)
		client.SendPayload(hwid, hostname, "software", collector.CollectSoftware(), false)
		client.SendPayload(hwid, hostname, "usb", collector.CollectUSB(), false)
	}
}
