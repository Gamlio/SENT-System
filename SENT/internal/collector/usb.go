package collector

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

// Các biểu thức chính quy dùng chung[cite: 92]
var (
	reWindowsVID = regexp.MustCompile(`VID_([A-Za-z0-9]{4})`)
	reWindowsPID = regexp.MustCompile(`PID_([A-Za-z0-9]{4})`)
)

// USBSensor định nghĩa sensor thu thập thiết bị ngoại vi[cite: 92]
type USBSensor struct{}

// USBRecord định nghĩa cấu trúc dữ liệu gửi về Backend[cite: 92]
type USBRecord struct {
	DeviceName   string `json:"device_name"`
	DeviceID     string `json:"device_id"`
	VID          string `json:"vid"`
	PID          string `json:"pid"`
	SerialNumber string `json:"serial_number"`
	DeviceHash   string `json:"device_hash"`
	EventType    string `json:"event_type"`
}

func (s *USBSensor) Name() string {
	return "usb"
}

// --- CÁC HÀM PHỤ TRỢ DÙNG CHUNG ---

// generateDeviceHash tạo mã băm định danh duy nhất cho thiết bị[cite: 92]
func generateDeviceHash(vid, pid, serial string) string {
	rawString := fmt.Sprintf("%s|%s|%s", vid, pid, serial)
	hash := sha256.Sum256([]byte(rawString))
	return hex.EncodeToString(hash[:])
}

// parseWindowsUSBID hỗ trợ bóc tách ID cho cả Windows và Linux SysFS[cite: 92]
func parseWindowsUSBID(pnpID string) (vid, pid, serial string) {
	vid, pid, serial = "UNKNOWN", "UNKNOWN", "UNKNOWN"

	if match := reWindowsVID.FindStringSubmatch(pnpID); len(match) > 1 {
		vid = match[1]
	}
	if match := reWindowsPID.FindStringSubmatch(pnpID); len(match) > 1 {
		pid = match[1]
	}

	parts := strings.Split(pnpID, "\\")
	if len(parts) >= 3 {
		serial = parts[len(parts)-1]
	}

	return vid, pid, serial
}
