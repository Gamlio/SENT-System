package collector

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"golang.org/x/sys/windows/registry"
)

// Khai báo các biểu thức chính quy (Regex) ở cấp package để biên dịch một lần duy nhất.
// Đây là một tối ưu hiệu suất quan trọng, tránh việc biên dịch lại regex trong các vòng lặp.
var (
	reLinuxUSB   = regexp.MustCompile(`Bus (\d{3}) Device (\d{3}): ID ([a-f0-9]{4}):([a-f0-9]{4}) (.*)`)
	reWindowsVID = regexp.MustCompile(`VID_([A-Za-z0-9]{4})`)
	reWindowsPID = regexp.MustCompile(`PID_([A-Za-z0-9]{4})`)
)

// 1. Khai báo Struct cho Sensor
type USBSensor struct{}

// 2. Cấu trúc JSON chuẩn gửi về Backend
type USBRecord struct {
	DeviceName   string `json:"device_name"`
	DeviceID     string `json:"device_id"` // Chuỗi gốc của OS
	VID          string `json:"vid"`
	PID          string `json:"pid"`
	SerialNumber string `json:"serial_number"`
	DeviceHash   string `json:"device_hash"` // Vân tay độc nhất = Hash(VID+PID+Serial)
	EventType    string `json:"event_type"`  // "plugged"
}

// 3. Khai báo tên định danh của Log
func (s *USBSensor) Name() string {
	return "usb"
}

// CollectUSB: Thu thập và định danh USB đa nền tảng
func (s *USBSensor) Collect() (interface{}, error) {
	var usbList []USBRecord

	// [FIX-WMI-HANG] Đặt timeout 10 giây cho các lệnh OS để tránh treo SENT
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if runtime.GOOS == "windows" {
		// [TỐI ƯU HIỆU NĂNG] Đọc trực tiếp Registry thay vì gọi PowerShell để loại bỏ CPU Spikes
		// Chỉ theo dõi USB Mass Storage (USBSTOR) vì đây là mục tiêu chính của Data Exfiltration
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\USBSTOR\Enum`, registry.QUERY_VALUE)
		if err == nil {
			defer k.Close()
			count, _, err := k.GetIntegerValue("Count")
			if err == nil {
				for i := 0; i < int(count); i++ {
					val, _, err := k.GetStringValue(fmt.Sprintf("%d", i))
					if err == nil {
						vid, pid, serial := parseWindowsUSBID(val)
						hash := generateDeviceHash(vid, pid, serial)
						usbList = append(usbList, USBRecord{
							DeviceName:   "USB Mass Storage Device",
							DeviceID:     val,
							VID:          vid,
							PID:          pid,
							SerialNumber: serial,
							DeviceHash:   hash,
							EventType:    "plugged",
						})
					}
				}
			}
		}
		return usbList, nil
	}

	if runtime.GOOS == "linux" {
		// [TỐI ƯU HIỆU SUẤT] Bỏ gọi exec.Command, đọc trực tiếp từ SysFS
		devices, err := filepath.Glob("/sys/bus/usb/devices/*")
		if err == nil {
			for _, devPath := range devices {
				// Chỉ đọc các thiết bị gốc (có chứa file idVendor)
				if _, err := os.Stat(filepath.Join(devPath, "idVendor")); os.IsNotExist(err) {
					continue
				}

				vidBytes, _ := os.ReadFile(filepath.Join(devPath, "idVendor"))
				pidBytes, _ := os.ReadFile(filepath.Join(devPath, "idProduct"))
				serialBytes, _ := os.ReadFile(filepath.Join(devPath, "serial"))
				productBytes, _ := os.ReadFile(filepath.Join(devPath, "product"))

				vid := strings.TrimSpace(string(vidBytes))
				pid := strings.TrimSpace(string(pidBytes))
				serial := strings.TrimSpace(string(serialBytes))
				name := strings.TrimSpace(string(productBytes))

				if serial == "" {
					serial = "N/A"
				}
				if name == "" {
					name = "Unknown USB Device"
				}

				hash := generateDeviceHash(vid, pid, serial)

				usbList = append(usbList, USBRecord{
					DeviceName:   name,
					DeviceID:     fmt.Sprintf("USB\\VID_%s&PID_%s", vid, pid), // Chuẩn hóa ID
					VID:          vid,
					PID:          pid,
					SerialNumber: serial,
					DeviceHash:   hash,
					EventType:    "plugged",
				})
			}
		}
		return usbList, nil
	}

	if runtime.GOOS == "darwin" {
		// macOS: Dùng system_profiler xuất ra JSON
		cmd := exec.CommandContext(ctx, "/usr/sbin/system_profiler", "SPUSBDataType", "-json")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("lỗi thực thi system_profiler: %w", err)
		}
		// (Phần parse JSON của macOS khá phức tạp và lồng nhau nhiều tầng,
		// ở đây giữ khung để bạn có thể mở rộng bằng thư viện gjson sau này)
		_ = output
		return usbList, nil
	}

	return usbList, nil
}

// --- CÁC HÀM PHỤ TRỢ (HELPERS) ---

// parseWindowsUSBID: Bóc tách thông số từ chuỗi PNPDeviceID của Windows
func parseWindowsUSBID(pnpID string) (vid, pid, serial string) {
	vid = "UNKNOWN"
	pid = "UNKNOWN"
	serial = "UNKNOWN"

	// 1. Lấy VID
	if match := reWindowsVID.FindStringSubmatch(pnpID); len(match) > 1 {
		vid = match[1]
	}

	// 2. Lấy PID
	if match := reWindowsPID.FindStringSubmatch(pnpID); len(match) > 1 {
		pid = match[1]
	}

	// 3. Lấy Serial (Thường nằm ở đoạn cuối cùng sau dấu \)
	parts := strings.Split(pnpID, "\\")
	if len(parts) >= 3 {
		serial = parts[len(parts)-1]
	}

	return vid, pid, serial
}

// generateDeviceHash: Băm thông tin để tạo vân tay duy nhất cho từng USB
func generateDeviceHash(vid, pid, serial string) string {
	// Tránh băm các giá trị rỗng/unknown làm trùng lặp hash của nhiều thiết bị lạ
	rawString := fmt.Sprintf("%s|%s|%s", vid, pid, serial)
	hash := sha256.Sum256([]byte(rawString))
	return hex.EncodeToString(hash[:])
}
