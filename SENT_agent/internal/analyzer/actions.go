package analyzer

import (
	"fmt"
	"os/exec"
	"runtime"
)

// ShowSystemAlert: Hiển thị Pop-up cảnh báo trên Windows, macOS và Linux
func ShowSystemAlert(title, message string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		psScript := fmt.Sprintf(`Add-Type -AssemblyName PresentationFramework; [System.Windows.MessageBox]::Show('%s', '%s', 'OK', 'Warning')`, message, title)
		cmd = exec.Command("powershell", "-WindowStyle", "Hidden", "-Command", psScript)
	case "darwin": // macOS
		appleScript := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
		cmd = exec.Command("osascript", "-e", appleScript)
	case "linux":
		// Sử dụng notify-send phổ biến trên Ubuntu/Debian
		cmd = exec.Command("notify-send", title, message, "-u", "critical")
	default:
		fmt.Printf("⚠️ [%s] %s\n", title, message)
		return
	}

	err := cmd.Start()
	if err != nil {
		fmt.Printf("⚠️ Lỗi hiển thị cảnh báo UI: %v\n", err)
	}
}
