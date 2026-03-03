package collector

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Sửa tên hàm: collectUSB -> CollectUSB (Để main.go có thể gọi được)
func CollectUSB() interface{} {
	if runtime.GOOS == "windows" {
		psCommand := `
			$usb = @(Get-CimInstance Win32_PnPEntity | Where-Object { $_.DeviceID -match '^USB\\' -and $_.Status -eq 'OK' })
			if ($usb.Count -eq 0) { Write-Output "[]"; exit }
			$usb | Select-Object Name, DeviceID | ConvertTo-Json -Compress
		`
		cmd := exec.Command("powershell", "-NoProfile", "-Command", psCommand)
		output, err := cmd.Output()
		if err != nil {
			return []map[string]interface{}{}
		}

		jsonStr := strings.TrimSpace(string(output))
		if len(jsonStr) > 0 && jsonStr[0] != '[' {
			jsonStr = "[" + jsonStr + "]"
		}

		var rawList []map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &rawList); err != nil {
			return []map[string]interface{}{}
		}

		var usbList []map[string]interface{}
		for _, dev := range rawList {
			name := fmt.Sprintf("%v", dev["Name"])
			id := fmt.Sprintf("%v", dev["DeviceID"])
			usbList = append(usbList, map[string]interface{}{
				"device_name": name,
				"device_id":   id,
				"event_type":  "plugged",
			})
		}
		return usbList
	}

	if runtime.GOOS == "linux" {
		cmd := exec.Command("lsusb")
		output, err := cmd.Output()
		if err != nil {
			return []map[string]interface{}{}
		}

		var usbList []map[string]interface{}
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			usbList = append(usbList, map[string]interface{}{
				"device_name": line,
				"device_id":   line,
				"event_type":  "plugged",
			})
		}
		return usbList
	}
	return []map[string]interface{}{}
}
