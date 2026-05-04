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

	// Tạo các thư mục cần thiết nếu chưa có
	os.MkdirAll("data", 0755)
	os.MkdirAll("config", 0755)

	err := collector.InitLocalDB("data")
	if err != nil {
		log.Fatalf("Không thể khởi tạo DB: %v", err)
	}
	defer collector.CloseLocalDB() // Đảm bảo nhả file agent_cache.db.lock khi thoát

	err = godotenv.Load("config/.env")
	if err != nil {
		log.Println("⚠️  Cảnh báo: Không tìm thấy file .env trong config/, sử dụng cấu hình mặc định.")
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
	immediateTicker := time.NewTicker(5 * time.Second) // Tức khắc (USB)
	fiveMinTicker := time.NewTicker(5 * time.Minute)   // Nhóm 5 phút (Port, Data Transfer, AV, Firewall)
	hourlyTicker := time.NewTicker(1 * time.Hour)      // Nhóm 1 tiếng (Software)
	tenHourTicker := time.NewTicker(10 * time.Hour)    // Nhóm 5-10 tiếng (Inventory)

	// Gửi toàn bộ trạng thái Baseline ngay lần đầu Agent khởi động
	runSensorsBatch(client, AssetHWID, hostname, "immediate")
	runSensorsBatch(client, AssetHWID, hostname, "five_min")
	runSensorsBatch(client, AssetHWID, hostname, "hourly")
	runSensorsBatch(client, AssetHWID, hostname, "ten_hour")

	for {
		select {
		case <-immediateTicker.C:
			runSensorsBatch(client, AssetHWID, hostname, "immediate")
		case <-fiveMinTicker.C:
			runSensorsBatch(client, AssetHWID, hostname, "five_min")
		case <-hourlyTicker.C:
			runSensorsBatch(client, AssetHWID, hostname, "hourly")
		case <-tenHourTicker.C:
			runSensorsBatch(client, AssetHWID, hostname, "ten_hour")
		}
	}
}

func runSensorsBatch(client *transport.AssetClient, hwid, hostname string, group string) {
	batchData := make(map[string]interface{})
	batchHashes := make(map[string]string)
	hasNewData := false

	for _, sensor := range collector.Registry {
		name := sensor.Name()

		// Phân loại nhóm chu kỳ quét
		var expectedGroup string
		switch name {
		case "usb":
			expectedGroup = "immediate"
		case "port", "data_transfer", "antivirus":
			expectedGroup = "five_min"
		case "software":
			expectedGroup = "hourly"
		case "inventory":
			expectedGroup = "ten_hour"
		default:
			expectedGroup = "five_min"
		}

		// Chỉ kích hoạt sensor nếu đúng chu kỳ của nhóm
		if expectedGroup != group {
			continue
		}

		data, err := sensor.Collect()
		if err == nil && data != nil {
			// 1. Tính toán vân tay (Hash) của bộ dữ liệu hiện tại
			currentHash := utils.CalculateHash(data)

			// 2. So sánh với bản cũ trong DB
			if collector.IsDataNew(sensor.Name(), currentHash) {
				batchData[name] = data
				batchHashes[name] = currentHash
				hasNewData = true
			}
		}
	}

	if hasNewData {
		go func(groupName string, payload interface{}, hashes map[string]string) {
			// 3. Gửi payload gom nhóm, nếu Server chưa phê duyệt (trả về lỗi) thì không lưu state
			status := client.SendPayload(hwid, hostname, "batch_"+groupName, payload)
			// 4. Chỉ lưu lại trạng thái mới để lần sau không gửi trùng khi Server đã lưu OK
			if status == "OK" {
				for modName, hash := range hashes {
					collector.UpdateModuleState(modName, hash)
				}
			}
		}(group, batchData, batchHashes)
	}
}
