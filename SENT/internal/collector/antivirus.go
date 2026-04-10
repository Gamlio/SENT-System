package collector

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/StackExchange/wmi"
)

// 1. Khai báo Struct cho Sensor
type AntivirusSensor struct{}

// 2. Cấu trúc JSON chuẩn gửi về Backend
type AntivirusRecord struct {
	HasThreat bool `json:"has_threat"`
}

// 3. Khai báo tên định danh của Log
func (s *AntivirusSensor) Name() string {
	return "antivirus"
}

// CollectAntivirus: Thu thập trạng thái mã độc / trình diệt virus
func (s *AntivirusSensor) Collect() (interface{}, error) {
	switch runtime.GOOS {
	case "windows":
		type MSFT_MpThreat struct {
			RollupStatus uint32
		}
		var threats []MSFT_MpThreat
		q := wmi.CreateQuery(&threats, "WHERE RollupStatus = 1")
		// Nếu truy vấn WMI lỗi (VD: Defender không chạy), mặc định là không có threat.
		if err := wmi.Query(q, &threats, nil, `ROOT\Microsoft\Windows\Defender`); err == nil {
			if len(threats) > 0 {
				return AntivirusRecord{HasThreat: true}, nil
			}
		}
		return AntivirusRecord{HasThreat: false}, nil

	case "darwin":
		cmd := exec.Command("spctl", "--status")
		out, err := cmd.Output()
		// Gatekeeper bị tắt là một rủi ro bảo mật.
		if err == nil && strings.Contains(strings.ToLower(string(out)), "disabled") {
			return AntivirusRecord{HasThreat: true}, nil
		}
		return AntivirusRecord{HasThreat: false}, nil

	case "linux":
		// Cảnh báo: Logic này quá đơn giản, chỉ check ClamAV.
		// Một hệ thống thực tế cần kiểm tra nhiều AV khác hoặc dùng API chuyên dụng.
		cmd := exec.Command("systemctl", "is-active", "clamav-daemon")
		out, err := cmd.Output()
		// Nếu service không "active" (bao gồm lỗi hoặc trạng thái inactive), coi như là một rủi ro.
		if err != nil || !strings.Contains(string(out), "active") {
			return AntivirusRecord{HasThreat: true}, nil
		}
		return AntivirusRecord{HasThreat: false}, nil

	default:
		return AntivirusRecord{HasThreat: false}, nil // Hệ điều hành không hỗ trợ
	}
}
