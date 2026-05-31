//go:build darwin

package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// SetupAutoStart cấu hình ứng dụng khởi động cùng MacOS bằng LaunchDaemon
func SetupAutoStart() {
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("⚠️ Lỗi lấy đường dẫn thực thi: %v", err)
		return
	}
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return
	}

	plistPath := "/Library/LaunchDaemons/com.sent.agent.plist"

	if _, err := os.Stat(plistPath); err == nil {
		return
	}

	workDir := filepath.Dir(exePath)

	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.sent.agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
        <key>SENT_DAEMON</key>
        <string>1</string>
    </dict>
    <key>WorkingDirectory</key>
    <string>%s</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
`, exePath, workDir)

	err = os.WriteFile(plistPath, []byte(plistContent), 0644)
	if err != nil {
		log.Printf("⚠️ Không thể tạo LaunchDaemon (yêu cầu quyền Root): %v", err)
		return
	}
	
	log.Println("✅ Đã thiết lập LaunchDaemon để khởi động cùng hệ thống trên macOS thành công.")
}
