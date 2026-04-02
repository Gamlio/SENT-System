package collector

import (
	"os/exec"
	"runtime"
	"strings"
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
	hasThreat := false
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		psCmd := `Get-MpThreat | Where-Object { $_.RollupStatus -eq 1 } | Select-Object -First 1`
		cmd = exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	case "darwin":
		cmd = exec.Command("spctl", "--status")
	case "linux":
		// Cảnh báo: Logic này quá đơn giản, chỉ check ClamAV.
		// Một hệ thống thực tế cần kiểm tra nhiều AV khác hoặc dùng API chuyên dụng.
		cmd = exec.Command("systemctl", "is-active", "clamav-daemon")
	default:
		return AntivirusRecord{HasThreat: false}, nil // Hệ điều hành không hỗ trợ
	}

	out, err := cmd.Output()
	// Lỗi có thể xảy ra nếu command không tồn tại (VD: spctl trên Linux), không phải lúc nào cũng là lỗi nghiêm trọng.
	// Ta chỉ quan tâm đến output, nên sẽ không trả về lỗi ở đây, nhưng trong hệ thống thực tế nên log lại.
	if err == nil {
		output := strings.ToLower(string(out))
		hasThreat = (runtime.GOOS == "windows" && len(strings.TrimSpace(output)) > 0) || (runtime.GOOS == "darwin" && strings.Contains(output, "disabled")) || (runtime.GOOS == "linux" && strings.Contains(output, "inactive"))
	}

	return AntivirusRecord{HasThreat: hasThreat}, nil
}
