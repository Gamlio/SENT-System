package analyzer

import (
	"fmt"
	"os/exec"
	"runtime"
	"sent_agent/internal/transport"
	"strings"
)

// Biến lưu trữ điểm rủi ro cục bộ của máy trạm
var LocalRiskScore = 0
var AlertCache = make(map[string]bool)

// ------------------------------------------------------------------
// USECASE 1: PHẦN MỀM TRÁI PHÉP (P3 - +20 Điểm)
// ------------------------------------------------------------------
func EvaluateSoftware(softwareData interface{}, blacklist []string) {
	// Ép kiểu từ interface{} (do collector trả về) sang dạng Slice map
	list, ok := softwareData.([]map[string]string)
	if !ok {
		return
	}

	for _, sw := range list {
		swName := sw["software_name"]
		for _, b := range blacklist {
			if strings.Contains(strings.ToLower(swName), strings.ToLower(b)) {
				fmt.Println(" [P3] Phát hiện phần mềm cấm:", swName)
				LocalRiskScore += 20

				ShowSystemAlert("CẢNH BÁO AN NINH (P3)",
					fmt.Sprintf("Phát hiện phần mềm không hợp lệ: %s.\nVui lòng gỡ cài đặt để tuân thủ chính sách công ty.", swName))
				return // Tránh spam nhiều popup cùng lúc
			}
		}
	}
}

// ------------------------------------------------------------------
// USECASE 2: USB TRÁI PHÉP (P3 - +20 Điểm)
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
			LocalRiskScore += 20 // Điểm rủi ro vẫn cộng để duy trì trạng thái nguy hiểm

			// [FIX SPAM] Chỉ hiện Popup và gửi Alert nếu chưa từng báo lỗi USB này
			if !AlertCache[devID] {
				fmt.Println(" [P3] Phát hiện USB lạ:", devName)
				ShowSystemAlert("CẢNH BÁO AN NINH (P3)", "Bạn vừa cắm thiết bị USB chưa được đăng ký.\nIT đang ghi nhận sự kiện này.")
				// transport.AgentClient.SendAlert(...) // (Nếu có hàm gửi Alert USB thì gọi ở đây)

				AlertCache[devID] = true // Đánh dấu là đã báo cáo
			}
			return
		}
	}
}

// ------------------------------------------------------------------
// USECASE 3: MỞ CỔNG MẠNG TRÁI PHÉP (P2 - +50 Điểm)
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
				fmt.Println(" [P2] Mở cổng trái phép:", portNum)
				LocalRiskScore += 50
				ShowSystemAlert("CẢNH BÁO NGUY HIỂM (P2)", fmt.Sprintf("Máy bạn đang mở port nhạy cảm (%d).\nHãy tắt ngay dịch vụ liên quan (VD: RDP, Telnet).", portNum))
				return
			}
		}
	}
}

// ------------------------------------------------------------------
// USECASE 4: KIỂM TRA TƯỜNG LỬA (FIREWALL) - ĐA NỀN TẢNG (P1)
// ------------------------------------------------------------------
func EvaluateFirewall(hwid, hostname string, data interface{}) {
	fwData, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	isOff, _ := fwData["firewall_off"].(bool)
	if isOff {
		fmt.Println(" [P1] TƯỜNG LỬA HỆ THỐNG ĐANG BỊ TẮT!")
		LocalRiskScore += 80
		ShowSystemAlert("CẢNH BÁO BẢO MẬT (P1)", "Tường lửa đã bị vô hiệu hóa. Vui lòng bật lại ngay!")
		transport.AgentClient.SendAlert(hwid, hostname, "Firewall Disabled", "Phát hiện Tường lửa bị tắt trên máy trạm.", "Critical")
	}
}
func CheckFirewallStatus(hwid, hostname string) {
	var isOff bool
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "netsh advfirewall show allprofiles state")
		out, _ := cmd.Output()
		isOff = strings.Contains(strings.ToLower(string(out)), "off")
	case "linux":
		// Kiểm tra UFW (Ubuntu/Debian)
		cmd = exec.Command("ufw", "status")
		out, _ := cmd.Output()
		isOff = strings.Contains(strings.ToLower(string(out)), "inactive")
	case "darwin":
		// Kiểm tra Application Firewall của macOS
		cmd = exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate")
		out, _ := cmd.Output()
		isOff = strings.Contains(strings.ToLower(string(out)), "disabled")
	}

	if isOff {
		fmt.Println("[P1] TƯỜNG LỬA HỆ THỐNG ĐANG BỊ TẮT!")
		LocalRiskScore += 80
		ShowSystemAlert("CẢNH BÁO BẢO MẬT (P1)", "Tường lửa đã bị vô hiệu hóa. Vui lòng bật lại ngay!")
		transport.AgentClient.SendAlert(hwid, hostname, "Firewall Disabled", "Phát hiện Tường lửa bị tắt trên máy trạm.", "Critical")
	}
}

// ------------------------------------------------------------------
// USECASE 5: KIỂM TRA MÃ ĐỘC / BẢO VỆ THỜI GIAN THỰC (P2)
// ------------------------------------------------------------------
func EvaluateAntivirus(hwid, hostname string, data interface{}) {
	avData, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	hasThreat, _ := avData["has_threat"].(bool)
	if hasThreat {
		fmt.Println(" [P2] PHÁT HIỆN RỦI RO MÃ ĐỘC / TẮT TRÌNH DIỆT VIRUS!")
		LocalRiskScore += 50
		ShowSystemAlert("CẢNH BÁO MÃ ĐỘC (P2)", "Hệ thống bảo vệ (Antivirus/Gatekeeper) đang báo động hoặc bị vô hiệu hóa!")
		transport.AgentClient.SendAlert(hwid, hostname, "Malware/AV Alert", "Phát hiện mã độc hoặc phần mềm diệt Virus bị tắt.", "High")
	}
}
func CheckAntivirusStatus(hwid, hostname string) {
	var hasThreat bool
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Quét xem Windows Defender có phát hiện mã độc đang active không
		psCmd := `Get-MpThreat | Where-Object { $_.RollupStatus -eq 1 } | Select-Object -First 1`
		cmd = exec.Command("powershell", "-NoProfile", "-Command", psCmd)
		out, _ := cmd.Output()
		hasThreat = len(strings.TrimSpace(string(out))) > 0
	case "darwin":
		// Kiểm tra Gatekeeper của macOS có bị tắt không (Hành vi nguy hiểm)
		cmd = exec.Command("spctl", "--status")
		out, _ := cmd.Output()
		hasThreat = strings.Contains(strings.ToLower(string(out)), "disabled")
	case "linux":
		// Giả định dùng ClamAV, nếu daemon không chạy -> Rủi ro
		cmd = exec.Command("systemctl", "is-active", "clamav-daemon")
		out, _ := cmd.Output()
		hasThreat = strings.Contains(strings.ToLower(string(out)), "inactive")
	}

	if hasThreat {
		LocalRiskScore += 50

		// [FIX SPAM]
		if !AlertCache["malware_active"] {
			fmt.Println("[P2] PHÁT HIỆN MÃ ĐỘC!")
			ShowSystemAlert("CẢNH BÁO MÃ ĐỘC (P2)", "Windows Defender báo cáo máy bạn đang nhiễm mã độc. IT sẽ liên hệ hỗ trợ.")
			transport.AgentClient.SendAlert(hwid, hostname, "Malware/AV Alert", "Phát hiện mã độc thông qua Windows Defender chưa được xử lý.", "High")

			AlertCache["malware_active"] = true
		}
	} else {
		// Nếu đã quét sạch virus thì reset lại cache để lần sau bị lại thì còn báo
		AlertCache["malware_active"] = false
	}
}
