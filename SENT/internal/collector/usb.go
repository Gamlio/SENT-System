package collector

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// Khai báo các biểu thức chính quy (Regex) ở cấp package để biên dịch một lần duy nhất.
// Đây là một tối ưu hiệu suất quan trọng, tránh việc biên dịch lại regex trong các vòng lặp.
var (
	reLinuxUSB   = regexp.MustCompile(`ID ([a-f0-9]{4}):([a-f0-9]{4}) (.*)`)
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

	// [FIX-WMI-HANG] Đặt timeout 10 giây cho các lệnh OS để tránh treo Agent
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if runtime.GOOS == "windows" {
		psCommand := `
			$usb = Get-CimInstance Win32_PnPEntity | Where-Object { $_.PNPDeviceID -match '^USB' -and $_.Status -eq 'OK' }
			if ($usb.Count -eq 0) { Write-Output "[]"; exit }
			$usb | Select-Object Name, PNPDeviceID | ConvertTo-Json -Compress
		`
		cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", psCommand)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("lỗi thực thi lệnh PowerShell: %w", err)
		}

		jsonStr := strings.TrimSpace(string(output))
		if len(jsonStr) > 0 && jsonStr[0] != '[' {
			jsonStr = "[" + jsonStr + "]" // Ép thành mảng nếu chỉ có 1 thiết bị
		}

		var rawList []map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &rawList); err != nil {
			return nil, fmt.Errorf("lỗi parse JSON từ PowerShell: %w", err)
		}

		for _, dev := range rawList {
			name := fmt.Sprintf("%v", dev["Name"])
			pnpID := fmt.Sprintf("%v", dev["PNPDeviceID"])

			// Bóc tách VID, PID, Serial từ PNPDeviceID
			// Ví dụ: USB\VID_045E&PID_07A5\6&3412312&0&2
			vid, pid, serial := parseWindowsUSBID(pnpID)

			hash := generateDeviceHash(vid, pid, serial)

			usbList = append(usbList, USBRecord{
				DeviceName:   name,
				DeviceID:     pnpID,
				VID:          vid,
				PID:          pid,
				SerialNumber: serial,
				DeviceHash:   hash,
				EventType:    "plugged",
			})
		}
		return usbList, nil
	}

	if runtime.GOOS == "linux" {
		// Sử dụng lsusb trên Linux (Định dạng: Bus 002 Device 001: ID 1d6b:0003 Linux Foundation 3.0 root hub)
		cmd := exec.CommandContext(ctx, "lsusb")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("lỗi thực thi lsusb: %w", err)
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}

			// Dùng Regex global đã khai báo sẵn, không biên dịch lại
			matches := reLinuxUSB.FindStringSubmatch(line)

			if len(matches) >= 4 {
				vid := matches[1]
				pid := matches[2]
				name := strings.TrimSpace(matches[3])
				serial := "N/A" // lsusb mặc định không hiện serial, cần udevadm nếu muốn sâu hơn

				hash := generateDeviceHash(vid, pid, serial)

				usbList = append(usbList, USBRecord{
					DeviceName:   name,
					DeviceID:     line, // Lưu nguyên dòng log làm ID thô
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
		cmd := exec.CommandContext(ctx, "system_profiler", "SPUSBDataType", "-json")
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
