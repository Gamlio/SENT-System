package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/host"

	"SENT/internal/collector"
	"SENT/internal/config"
	"SENT/internal/transport"

	"SENT/internal/utils"
)

func main() {
	// Bắt lỗi crash (panic) và dừng màn hình để người dùng kịp đọc lỗi
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("\n🚨 [LỖI NGHIÊM TRỌNG]: %v\n", r)
		}
		log.Println("SENT đang tự động khởi động lại hoặc thoát do lỗi...")
	}()

	// [QUAN TRỌNG] Kiểm tra quyền Admin/Root trước khi làm bất cứ điều gì
	if !utils.IsAdmin() {
		log.Fatal("FATAL: asset yêu cầu quyền Administrator/Root để hoạt động. Vui lòng chạy lại bằng 'Run as Administrator' hoặc 'sudo'.")
	}

	hInfo, _ := host.Info()

	AssetHWID := hInfo.HostID

	hostname, _ := os.Hostname()
	hostname = fmt.Sprintf("%s-Virtual", hostname)

	fmt.Printf("\n 🛡️ SENT asset V4.0 (Ninja Thin-Client) | HOST: %s\n", hostname)
	ipAddress := utils.GetOutboundIP()
	config.LoadOrBootstrap(AssetHWID, hostname, ipAddress)

	// Từ đoạn này trở xuống giữ nguyên...
	client := transport.GetAssetClient()

	// 2. Khởi tạo danh sách các module thu thập (Plugins)
	collector.InitCollectors()
	fmt.Printf(" [INFO] Đã nạp %d module cảm biến...\n", len(collector.Registry))

	// Tách chu kỳ quét để tối ưu I/O Disk & CPU
	fastTicker := time.NewTicker(30 * time.Second) // Các module nhẹ (Network, USB, AV)
	slowTicker := time.NewTicker(5 * time.Minute)  // Các module nặng (Software Hash, Inventory)

	for {
		select {
		case <-fastTicker.C:
			runSensors(client, AssetHWID, hostname, false)
		case <-slowTicker.C:
			runSensors(client, AssetHWID, hostname, true)
		}
	}
}

func runSensors(client *transport.AssetClient, hwid, hostname string, runHeavy bool) {
	for _, sensor := range collector.Registry {
		isHeavy := sensor.Name() == "software" || sensor.Name() == "inventory"
		if isHeavy && !runHeavy {
			continue // Bỏ qua quét nặng nếu không tới chu kỳ
		}

		data, err := sensor.Collect()
		if err == nil && data != nil {
			go func(logType string, logData interface{}) {
				client.SendPayload(hwid, hostname, logType, logData)
			}(sensor.Name(), data)
		}
	}
}
