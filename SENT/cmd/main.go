package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
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

	defer collector.CloseLocalDB() // Đảm bảo nhả file agent_cache.db.lock khi thoát

	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Cảnh báo: Không tìm thấy file .env, sử dụng cấu hình mặc định.")
	}

	if !utils.IsAdmin() {
		log.Fatal("FATAL: asset yêu cầu quyền Administrator/Root để hoạt động. Vui lòng chạy lại bằng 'Run as Administrator' hoặc 'sudo'.")
	}

	hInfo, _ := host.Info()

	AssetHWID := hInfo.HostID

	hostname, _ := os.Hostname()

	// 2. Kiểm tra tính toàn vẹn của chính file chạy
	exePath, err := os.Executable()
	if err == nil {
		hash, err := utils.CalculateSHA256(exePath) // Cần expose hàm calculateSHA256 ra utils
		if err == nil {
			fmt.Printf(" [SEC] Mã toàn vẹn của SENT Agent: %s\n", hash)
			// TODO: Ở bản cập nhật Backend tới, có thể gửi mã hash này lên để SOC xác minh
		}
	}

	fmt.Printf("\n 🛡️ SENT asset V4.0 (Ninja Thin-Client) | HOST: %s\n", hostname)
	ipAddress := utils.GetOutboundIP()
	config.LoadOrBootstrap(AssetHWID, hostname, ipAddress)

	// Từ đoạn này trở xuống giữ nguyên...
	client := transport.GetAssetClient()

	// 2. Khởi tạo danh sách các module thu thập (Plugins)
	collector.InitCollectors()
	fmt.Printf(" [INFO] Đã nạp %d module cảm biến...\n", len(collector.Registry))

	// Tách chu kỳ quét để tối ưu I/O Disk & CPU
	usbTicker := time.NewTicker(5 * time.Second)   // Tạm thời Polling nhanh cho USB (chờ nâng cấp Event-driven)
	fastTicker := time.NewTicker(30 * time.Second) // Các module nhẹ (Network, AV, Firewall)
	slowTicker := time.NewTicker(5 * time.Minute)  // Các module nặng (Software Hash)
	hourlyTicker := time.NewTicker(1 * time.Hour)  // Inventory (Ít thay đổi)

	// Gửi toàn bộ trạng thái Baseline ngay lần đầu Agent khởi động
	runSensors(client, AssetHWID, hostname, "inventory")
	runSensors(client, AssetHWID, hostname, "software")
	runSensors(client, AssetHWID, hostname, "fast")
	runSensors(client, AssetHWID, hostname, "usb")

	for {
		select {
		case <-usbTicker.C:
			runSensors(client, AssetHWID, hostname, "usb")
		case <-fastTicker.C:
			runSensors(client, AssetHWID, hostname, "fast")
		case <-slowTicker.C:
			runSensors(client, AssetHWID, hostname, "software")
		case <-hourlyTicker.C:
			runSensors(client, AssetHWID, hostname, "inventory")
		}
	}
}

func runSensors(client *transport.AssetClient, hwid, hostname string, group string) {
	for _, sensor := range collector.Registry {
		name := sensor.Name()

		// Phân loại nhóm chu kỳ quét
		var expectedGroup string
		switch name {
		case "inventory":
			expectedGroup = "inventory"
		case "software":
			expectedGroup = "software"
		case "usb":
			expectedGroup = "usb"
		default:
			// Các module còn lại (port, data_transfer, antivirus, firewall)
			expectedGroup = "fast"
		}

		// Chỉ kích hoạt sensor nếu đúng chu kỳ của nhóm
		if expectedGroup != group {
			continue
		}

		data, err := sensor.Collect()
		if err == nil && data != nil {
			go func(logType string, logData interface{}) {
				client.SendPayload(hwid, hostname, logType, logData)
			}(sensor.Name(), data)
		}
	}
}
