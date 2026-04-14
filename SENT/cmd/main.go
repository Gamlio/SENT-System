package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/host"

	// Bổ sung import config, utils, và transport
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
		log.Println("Agent đang tự động khởi động lại hoặc thoát do lỗi...")
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

	// =========================================================================
	// [QUAN TRỌNG] ĐÂY LÀ CHỖ BẠN BỊ THIẾU
	// Phải gọi hàm này để nó dừng lại bắt nhập Token nếu chưa có file config!
	// =========================================================================
	ipAddress := utils.GetOutboundIP()
	config.LoadOrBootstrap(AssetHWID, hostname, ipAddress)

	// Từ đoạn này trở xuống giữ nguyên...
	client := transport.GetAssetClient()

	// [QUAN TRỌNG]: Bật kênh nhận lệnh WebSocket
	client.StartHybridCommunication(AssetHWID, func(cmdType string, cmdData interface{}) {
		fmt.Printf("🎯 [LỆNH] Nhận yêu cầu: %s\n", cmdType)

		switch cmdType {
		case "TRIGGER_BASELINE":
			fmt.Println("🔄 Lệnh hệ thống: Đang thiết lập lại Baseline toàn diện...")

			// Duyệt qua tất cả Sensor đã đăng ký trong Registry
			for _, sensor := range collector.Registry {
				// Lấy dữ liệu mới nhất từ Sensor
				sensorData, err := sensor.Collect()
				if err != nil {
					fmt.Printf("❌ Lỗi thu thập Baseline cho %s: %v\n", sensor.Name(), err)
					continue
				}

				// Gửi dữ liệu dưới dạng Baseline (_baseline)
				client.SendBaseline(AssetHWID, hostname, sensor.Name(), sensorData)
				fmt.Printf("✅ Đã cập nhật Baseline cho module: %s\n", sensor.Name())
			}

			fmt.Println("🚀 Hoàn tất đồng bộ Baseline Zero Trust!")
		case "ISOLATE":
			// Logic cô lập máy bằng cách tắt toàn bộ kết nối mạng
			fmt.Println("🛑 Đã nhận lệnh cô lập máy!")
		}
	})

	// 2. Khởi tạo danh sách các module thu thập (Plugins)
	collector.InitCollectors()
	fmt.Printf(" [INFO] Đã nạp %d module cảm biến...\n", len(collector.Registry))

	// Chu kỳ quét mặc định 30 giây
	ticker := time.NewTicker(30 * time.Second)

	for range ticker.C {
		fmt.Printf("\n[+] Đang quét hệ thống lúc: %s\n", time.Now().Format("15:04:05"))

		// 3. DUYỆT QUA TẤT CẢ CÁC SENSOR ĐÃ ĐĂNG KÝ
		for _, sensor := range collector.Registry {
			data, err := sensor.Collect()
			if err != nil {
				log.Printf("⚠️ [WARNING] Sensor '%s' gặp lỗi: %v", sensor.Name(), err)
				// Nếu data là nil do lỗi nghiêm trọng, bỏ qua không gửi để tránh lỗi Backend
				if data == nil {
					continue
				}
			}

			go func(logType string, logData interface{}) {
				status := client.SendPayload(AssetHWID, hostname, logType, logData, false)
				if status == "ISOLATED" {
					fmt.Println("🚨 [CẢNH BÁO] Máy trạm đã bị Server đưa vào khu vực Cách ly mạng!")
				}
			}(sensor.Name(), data)
		}
	}
}
