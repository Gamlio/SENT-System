package analyzer

import (
	"fmt"
	"net"
	"os/exec"
	"sent_agent/internal/transport"
	"strings"
)

// Biến lưu trữ điểm rủi ro cục bộ của máy trạm
var LocalRiskScore = 0

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
				fmt.Println("🚨 [P3] Phát hiện phần mềm cấm:", swName)
				LocalRiskScore += 20

				ShowWindowsAlert("CẢNH BÁO AN NINH (P3)",
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
			fmt.Println("🚨 [P3] Phát hiện USB lạ:", devName)
			LocalRiskScore += 20
			ShowWindowsAlert("CẢNH BÁO AN NINH (P3)", "Bạn vừa cắm thiết bị USB chưa được đăng ký.\nIT đang ghi nhận sự kiện này.")
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
				fmt.Println("🚨 [P2] Mở cổng trái phép:", portNum)
				LocalRiskScore += 50
				ShowWindowsAlert("CẢNH BÁO NGUY HIỂM (P2)", fmt.Sprintf("Máy bạn đang mở port nhạy cảm (%d).\nHãy tắt ngay dịch vụ liên quan (VD: RDP, Telnet).", portNum))
				return
			}
		}
	}
}

// ------------------------------------------------------------------
// USECASE 4: MULTIPLE NETWORK / DUAL HOMING (P1 - +80 Điểm)
// ------------------------------------------------------------------
func EvaluateDualHoming(hwid, hostname string) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return
	}

	activeNetworks := 0
	for _, i := range interfaces {
		// Bỏ qua loopback và các card mạng ảo/tắt
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// Nếu là IPv4 thực tế
			if ip != nil && ip.To4() != nil && !strings.HasPrefix(ip.String(), "169.254") {
				activeNetworks++
				break // Tính 1 IP cho 1 Card mạng
			}
		}
	}

	// Nếu dùng 2 mạng cùng lúc (VD: Vừa cắm dây mạng công ty, vừa bắt Wifi 4G)
	if activeNetworks > 1 {
		fmt.Println("🚨 [P1] VI PHẠM DUAL HOMING!")
		LocalRiskScore += 80
		ShowWindowsAlert("VI PHẠM NGHIÊM TRỌNG (P1)", "Phát hiện máy tính đang kết nối 2 mạng cùng lúc. Vui lòng ngắt mạng cá nhân!")

		// BẮN CẢNH BÁO KHẨN VỀ SERVER
		transport.AgentClient.SendAlert(hwid, hostname, "Dual Homing", "Phát hiện máy tính dùng 2 mạng cùng lúc (Nguy cơ Data Exfiltration)", "Critical")
	}
}

// USECASE 5: NHIỄM VIRUS (P2)
// ------------------------------------------------------------------
func EvaluateAntivirus(hwid, hostname string) {
	psCmd := `Get-MpThreat | Where-Object { $_.RollupStatus -eq 1 } | Select-Object -First 1`
	out, err := exec.Command("powershell", "-NoProfile", "-Command", psCmd).Output()

	if err == nil && len(strings.TrimSpace(string(out))) > 0 {
		fmt.Println("🚨 [P2] PHÁT HIỆN MÃ ĐỘC!")
		LocalRiskScore += 50
		ShowWindowsAlert("CẢNH BÁO MÃ ĐỘC (P2)", "Windows Defender báo cáo máy bạn đang nhiễm mã độc. IT sẽ liên hệ hỗ trợ.")

		// BẮN CẢNH BÁO KHẨN VỀ SERVER
		transport.AgentClient.SendAlert(hwid, hostname, "Malware Infection", "Phát hiện mã độc thông qua Windows Defender chưa được xử lý.", "High")
	}
}
