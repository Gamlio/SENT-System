//go:build linux

package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

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

	serviceName := "sent-agent.service"
	servicePath := fmt.Sprintf("/etc/systemd/system/%s", serviceName)

	if _, err := os.Stat(servicePath); err == nil {
		return
	}

	workDir := filepath.Dir(exePath)

	serviceContent := fmt.Sprintf(`[Unit]
Description=SENT Security Agent
After=network.target

[Service]
Type=simple
ExecStart=%s
WorkingDirectory=%s
Restart=always
RestartSec=10
Environment="SENT_DAEMON=1"

[Install]
WantedBy=multi-user.target
`, exePath, workDir)

	err = os.WriteFile(servicePath, []byte(serviceContent), 0644)
	if err != nil {
		log.Printf("⚠️ Không thể tạo Systemd service (yêu cầu quyền Root): %v", err)
		return
	}

	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", serviceName).Run()
	log.Println("✅ Đã thiết lập Systemd service để khởi động cùng hệ thống trên Linux thành công.")
}
