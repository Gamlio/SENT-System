package main

import (
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/host"

	"sent_agent/internal/collector"
	"sent_agent/internal/config"
	"sent_agent/internal/transport"
)

func main() {
	// 1. Khởi tạo cấu hình
	if !config.Load() {
		config.PromptForCompanyCode()
	}

	hInfo, _ := host.Info()
	hwid := hInfo.HostID
	hostname, _ := os.Hostname()

	fmt.Printf("\n 🛡️ SENT AGENT V4.0 (Ninja Thin-Client) | COMPANY: %s | HOST: %s\n", config.Current.CompanyCode, hostname)

	client := transport.AgentClient

	// 2. Khởi tạo danh sách các module thu thập (Plugins)
	collector.InitCollectors()
	fmt.Printf(" [INFO] Đã nạp %d module cảm biến...\n", len(collector.Registry))

	// Chu kỳ quét mặc định 30 giây
	ticker := time.NewTicker(30 * time.Second)

	for range ticker.C {
		fmt.Printf("\n[+] Đang quét hệ thống lúc: %s\n", time.Now().Format("15:04:05"))

		// 3. DUYỆT QUA TẤT CẢ CÁC SENSOR ĐÃ ĐĂNG KÝ
		for _, sensor := range collector.Registry {

			// Thu thập dữ liệu
			data := sensor.Collect()

			// Gửi dữ liệu bất đồng bộ (goroutine) để không làm lag vòng lặp
			go func(logType string, logData interface{}) {

				// Hàm SendPayload sẽ tự động băm Hash và so sánh.
				// Nếu Hash không đổi -> Bắn Heartbeat. Nếu Hash đổi -> Bắn Data
				status := client.SendPayload(hwid, hostname, logType, logData, false)

				// Xử lý lệnh phản hồi từ Server (Cô lập, Ngắt kết nối...)
				if status == "ISOLATED" {
					fmt.Println("🚨 [CẢNH BÁO] Máy trạm đã bị Server đưa vào khu vực Cách ly mạng!")
					// TODO: Viết logic cắt mạng thật tại đây nếu cần
				}

			}(sensor.Name(), data)

		}
	}
}
