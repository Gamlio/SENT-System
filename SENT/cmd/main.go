package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/host"

	"SENT/internal/collector"
	"SENT/internal/config"
	"SENT/internal/policy"
	"SENT/internal/transport"
	"SENT/internal/utils"
)

func main() {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("\n🚨 [LỖI NGHIÊM TRỌNG]: %v\n", r)
		}
	}()

	os.MkdirAll("data", 0755)
	os.MkdirAll("config", 0755)

	if !utils.IsAdmin() {
		log.Fatal("FATAL: asset yêu cầu quyền Administrator/Root để hoạt động. Vui lòng chạy lại bằng 'Run as Administrator' hoặc 'sudo'.")
	}

	if os.Getenv("SENT_DAEMON") != "1" {
		utils.SetupAutoStart()
	}

	hInfo, _ := host.Info()
	AssetHWID := hInfo.HostID
	hostname, _ := os.Hostname()

	if os.Getenv("SENT_DAEMON") != "1" {
		exePath, err := os.Executable()
		if err == nil {
			hash, err := utils.CalculateSHA256(exePath)
			if err == nil {
				fmt.Printf(" [SEC] Mã toàn vẹn của SENT Agent: %s\n", hash)
			}
		}
	}

	if os.Getenv("SENT_DAEMON") != "1" || runtime.GOOS == "windows" {
		fmt.Printf("\n 🛡️ SENT asset V4.0 (Ninja Thin-Client) | HOST: %s\n", hostname)
	}
	ipAddress := utils.GetOutboundIP()
	if os.Getenv("SENT_DAEMON") != "1" || runtime.GOOS == "windows" {
		fmt.Printf(" [DEBUG] IP phát hiện được: '%s'\n", ipAddress)
	}

	config.LoadOrBootstrap(AssetHWID, hostname, ipAddress)

	if os.Getenv("SENT_DAEMON") != "1" || runtime.GOOS == "windows" {
		fmt.Println("🚀 Khởi tạo Agent thành công. Ứng dụng đang chuyển sang chế độ chạy ngầm...")
		time.Sleep(1 * time.Second)
	}

	forkLinuxDaemon()

	if runtime.GOOS == "windows" {
		hideConsoleWindow()
	}

	policy.StartSync()
	err := collector.InitLocalDB("data")
	if err != nil {
		log.Fatalf("Không thể khởi tạo DB: %v", err)
	}
	defer collector.CloseLocalDB()

	client := transport.GetAssetClient()

	err = collector.InitBufferBucket()
	if err != nil {
		log.Printf("⚠️ Không thể khởi tạo phân vùng Offline Buffer: %v", err)
	}

	// Goroutine ngầm kiểm tra và giải phóng dữ liệu vùng đệm ngoại tuyến
	go func(c *transport.AssetClient, hID, hName string) {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			offlineLogs, keys := collector.PopAllOfflineBuffer()
			if len(offlineLogs) > 0 {
				status := c.SendPayload(hID, hName, "batch_offline_flush", offlineLogs)
				if status == "OK" {
					collector.ClearFlushedLogs(keys)
				}
			}
		}
	}(client, AssetHWID, hostname)

	collector.InitCollectors()

	immediateTicker := time.NewTicker(5 * time.Second)
	fiveMinTicker := time.NewTicker(5 * time.Minute)
	hourlyTicker := time.NewTicker(1 * time.Hour)
	tenHourTicker := time.NewTicker(10 * time.Hour)

	runSensorsBatch(client, AssetHWID, hostname, "immediate")
	time.Sleep(500 * time.Millisecond)
	runSensorsBatch(client, AssetHWID, hostname, "five_min")
	time.Sleep(500 * time.Millisecond)
	runSensorsBatch(client, AssetHWID, hostname, "hourly")
	time.Sleep(500 * time.Millisecond)
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

		var expectedGroup string
		switch name {
		case "usb":
			expectedGroup = "immediate"
		case "port", "data_transfer", "antivirus":
			expectedGroup = "five_min"
		case "software", "patch":
			expectedGroup = "hourly"
		case "inventory":
			expectedGroup = "ten_hour"
		default:
			expectedGroup = "five_min"
		}

		if expectedGroup != group {
			continue
		}

		data, err := sensor.Collect()
		if err == nil && data != nil {
			currentHash := utils.CalculateHash(data)
			if collector.IsDataNew(sensor.Name(), currentHash) {
				batchData[name] = data
				batchHashes[name] = currentHash
				hasNewData = true
			}
		}
	}

	if hasNewData {
		go func(groupName string, payload interface{}, hashes map[string]string) {
			status := client.SendPayload(hwid, hostname, "batch_"+groupName, payload)
			if status == "OK" {
				for modName, hash := range hashes {
					collector.UpdateModuleState(modName, hash)
				}
			} else {
				_ = collector.PushToOfflineBuffer("batch_"+groupName, payload)
			}
		}(group, batchData, batchHashes)
	}
}
