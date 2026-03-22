package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/host"

	// Bổ sung import config và utils
	"sent_agent/internal/collector"
	"sent_agent/internal/config"
	"sent_agent/internal/transport"
	"sent_agent/internal/utils"
)

func main() {
	hInfo, _ := host.Info()

	// 1. Mẹo Random HWID cho môi trường Test (để giả lập nhiều máy)
	rand.Seed(time.Now().UnixNano())
	hwid := fmt.Sprintf("%s-TEST-%d", hInfo.HostID, rand.Intn(9999))

	hostname, _ := os.Hostname()
	hostname = fmt.Sprintf("%s-Virtual", hostname)

	fmt.Printf("\n 🛡️ SENT AGENT V4.0 (Ninja Thin-Client) | HOST: %s\n", hostname)

	// =========================================================================
	// [QUAN TRỌNG] ĐÂY LÀ CHỖ BẠN BỊ THIẾU
	// Phải gọi hàm này để nó dừng lại bắt nhập Token nếu chưa có file config!
	// =========================================================================
	ipAddress := utils.GetOutboundIP()
	config.LoadOrBootstrap(hwid, hostname, ipAddress)

	// Từ đoạn này trở xuống giữ nguyên...
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
			data := sensor.Collect()

			go func(logType string, logData interface{}) {
				status := client.SendPayload(hwid, hostname, logType, logData, false)
				if status == "ISOLATED" {
					fmt.Println("🚨 [CẢNH BÁO] Máy trạm đã bị Server đưa vào khu vực Cách ly mạng!")
				}
			}(sensor.Name(), data)
		}
	}
}
