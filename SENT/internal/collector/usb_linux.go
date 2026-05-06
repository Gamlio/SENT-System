//go:build linux

package collector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Collect: Thu thập và định danh USB trên Linux (Ubuntu, Debian...) qua SysFS
func (s *USBSensor) Collect() (interface{}, error) {
	var usbList []USBRecord

	// [TỐI ƯU HIỆU SUẤT] Đọc trực tiếp từ SysFS thay vì gọi lệnh lsusb
	devices, err := filepath.Glob("/sys/bus/usb/devices/*")
	if err != nil {
		return nil, fmt.Errorf("không thể truy cập danh sách thiết bị USB: %w", err)
	}

	for _, devPath := range devices {
		// Chỉ đọc các thiết bị gốc (có chứa file idVendor)
		if _, err := os.Stat(filepath.Join(devPath, "idVendor")); os.IsNotExist(err) {
			continue
		}

		// Đọc thông tin VID, PID, Serial và Product Name từ kernel
		vidBytes, _ := os.ReadFile(filepath.Join(devPath, "idVendor"))
		pidBytes, _ := os.ReadFile(filepath.Join(devPath, "idProduct"))
		serialBytes, _ := os.ReadFile(filepath.Join(devPath, "serial"))
		productBytes, _ := os.ReadFile(filepath.Join(devPath, "product"))

		vid := strings.TrimSpace(string(vidBytes))
		pid := strings.TrimSpace(string(pidBytes))
		serial := strings.TrimSpace(string(serialBytes))
		name := strings.TrimSpace(string(productBytes))

		// Xử lý giá trị mặc định nếu kernel không cung cấp[cite: 92]
		if serial == "" {
			serial = "N/A"
		}
		if name == "" {
			name = "Unknown USB Device"
		}

		// Tạo vân tay duy nhất cho thiết bị[cite: 92]
		hash := generateDeviceHash(vid, pid, serial)

		usbList = append(usbList, USBRecord{
			DeviceName:   name,
			DeviceID:     fmt.Sprintf("USB\\VID_%s&PID_%s", vid, pid),
			VID:          vid,
			PID:          pid,
			SerialNumber: serial,
			DeviceHash:   hash,
			EventType:    "plugged",
		})
	}

	return usbList, nil
}
