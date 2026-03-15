package analyzer

import (
	"fmt"
	"os/exec"
	"runtime"
	"sent_agent/internal/transport"
	"strings"
	"time"
)

var LocalRiskScore = 0

// [ANTI-SPAM CACHE] Lưu thời điểm hiển thị Popup cuối cùng
var AlertTimeCache = make(map[string]time.Time)

// shouldNotify: Chỉ cho phép hiện Popup nếu đã qua 12 tiếng kể từ lần trước
func shouldNotify(alertKey string) bool {
	lastTime, exists := AlertTimeCache[alertKey]
	if !exists {
		return true // Lần đầu tiên vi phạm
	}
	return time.Since(lastTime) > 12*time.Hour
}

func markNotified(alertKey string) {
	AlertTimeCache[alertKey] = time.Now()
}

// KillProcessSilent: Âm thầm tiêu diệt tiến trình không cho phép (Phản xạ ngầm)
func KillProcessSilent(processName string) {
	cleanName := strings.Split(strings.ToLower(processName), ".")[0]
	if runtime.GOOS == "windows" {
		exec.Command("taskkill", "/F", "/IM", cleanName+".exe").Run()
	} else {
		exec.Command("pkill", "-9", "-f", cleanName).Run()
	}
}

// ------------------------------------------------------------------
// 1. PHẦN MỀM TRÁI PHÉP (+20 Điểm)
// ------------------------------------------------------------------
func EvaluateSoftware(softwareData interface{}, blacklist []string) {
	list, ok := softwareData.([]map[string]string)
	if !ok {
		return
	}

	for _, sw := range list {
		swName := sw["software_name"]
		for _, b := range blacklist {
			if strings.Contains(strings.ToLower(swName), strings.ToLower(b)) {
				LocalRiskScore += 20
				alertKey := "SW_" + swName

				// [HÀNH ĐỘNG]: Luôn luôn kill tiến trình
				KillProcessSilent(swName)

				// [THÔNG BÁO]: Có kiểm soát (Max 2 lần/ngày)
				if shouldNotify(alertKey) {
					fmt.Println(" [P3] Đã đóng phần mềm cấm:", swName)
					ShowSystemAlert("BẢO VỆ CHỦ ĐỘNG (P3)", fmt.Sprintf("Phần mềm cấm [%s] vừa bị hệ thống buộc dừng.", swName))
					markNotified(alertKey)
				}
				return
			}
		}
	}
}

// ------------------------------------------------------------------
// 2. USB TRÁI PHÉP (+20 Điểm)
// ------------------------------------------------------------------
func EvaluateUSB(usbData interface{}, whitelist []string) {
	list, ok := usbData.([]map[string]interface{})
	if !ok {
		return
	}

	for _, dev := range list {
		devID := fmt.Sprintf("%v", dev["device_id"])
		devName := fmt.Sprintf("%v", dev["device_name"])

		isSafe := false
		for _, w := range whitelist {
			if strings.Contains(devID, w) {
				isSafe = true
				break
			}
		}

		if !isSafe {
			LocalRiskScore += 20
			alertKey := "USB_" + devID

			if shouldNotify(alertKey) {
				fmt.Println(" [P3] Phát hiện USB lạ:", devName)
				ShowSystemAlert("CẢNH BÁO AN NINH (P3)", "Bạn vừa cắm thiết bị USB chưa được phê duyệt.\nHành động đã được ghi log gửi về SOC.")
				markNotified(alertKey)
			}
			return
		}
	}
}

// ------------------------------------------------------------------
// 3. MỞ CỔNG MẠNG TRÁI PHÉP (+50 Điểm)
// ------------------------------------------------------------------
func EvaluateOpenPorts(telemetryData interface{}, blockedPorts []int) {
	data, ok := telemetryData.(map[string]interface{})
	if !ok {
		return
	}

	openPorts, ok := data["open_ports"].([]map[string]interface{})
	if !ok {
		return
	}

	for _, p := range openPorts {
		portNum, ok := p["port"].(int)
		if !ok {
			continue
		}

		for _, bp := range blockedPorts {
			if portNum == bp {
				LocalRiskScore += 50
				alertKey := fmt.Sprintf("PORT_%d", portNum)

				if shouldNotify(alertKey) {
					fmt.Println(" [P2] Mở cổng trái phép:", portNum)
					ShowSystemAlert("CẢNH BÁO NGUY HIỂM (P2)", fmt.Sprintf("Máy bạn đang mở cổng mạng nhạy cảm (%d).", portNum))
					markNotified(alertKey)
				}
				return
			}
		}
	}
}

// ------------------------------------------------------------------
// 4. KIỂM TRA TƯỜNG LỬA (+80 Điểm) -> TỰ ĐỘNG BẬT LẠI (Self-Healing)
// ------------------------------------------------------------------
func CheckFirewallStatus(hwid, hostname string) {
	var isOff bool
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "netsh advfirewall show allprofiles state")
		out, _ := cmd.Output()
		isOff = strings.Contains(strings.ToLower(string(out)), "off")
		// (Giữ nguyên macOS, Linux nếu cần)
	}

	if isOff {
		LocalRiskScore += 80
		alertKey := "FW_DISABLED"

		// [HÀNH ĐỘNG]: Tự động ép bật lại Firewall
		if runtime.GOOS == "windows" {
			exec.Command("netsh", "advfirewall", "set", "allprofiles", "state", "on").Run()
		}

		if shouldNotify(alertKey) {
			ShowSystemAlert("HỆ THỐNG TỰ PHỤC HỒI (P1)", "Tường lửa bị tắt trái phép. Hệ thống đã tự động bật lại để bảo vệ máy!")
			transport.AgentClient.SendAlert(hwid, hostname, "Firewall Disabled", "Tường lửa bị tắt. Agent đã tự động ép bật lại.", "Critical")
			markNotified(alertKey)
		}
	}
}

// ------------------------------------------------------------------
// 5. KIỂM TRA MÃ ĐỘC (+50 Điểm)
// ------------------------------------------------------------------
func CheckAntivirusStatus(hwid, hostname string) {
	var hasThreat bool
	if runtime.GOOS == "windows" {
		psCmd := `Get-MpThreat | Where-Object { $_.RollupStatus -eq 1 } | Select-Object -First 1`
		out, _ := exec.Command("powershell", "-NoProfile", "-Command", psCmd).Output()
		hasThreat = len(strings.TrimSpace(string(out))) > 0
	}

	if hasThreat {
		LocalRiskScore += 50
		alertKey := "MALWARE_DETECTED"

		if shouldNotify(alertKey) {
			ShowSystemAlert("CẢNH BÁO MÃ ĐỘC (P2)", "Phát hiện mã độc đang hoạt động. Vui lòng ngắt mạng và báo IT.")
			transport.AgentClient.SendAlert(hwid, hostname, "Malware/AV Alert", "Phát hiện mã độc qua Windows Defender.", "High")
			markNotified(alertKey)
		}
	}
}
