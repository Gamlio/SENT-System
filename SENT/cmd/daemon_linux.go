//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"SENT/internal/utils"
)

func forkLinuxDaemon() {
	if os.Getenv("SENT_DAEMON") != "1" {
		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		cmd.Env = append(os.Environ(), "SENT_DAEMON=1")
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}

		logFile, err := os.OpenFile("data/.sent_daemon.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err == nil {
			cmd.Stdout = logFile
			cmd.Stderr = logFile
		}

		if err := cmd.Start(); err != nil {
			fmt.Printf("❌ Lỗi fork tiến trình ngầm: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n✅ Cấu hình hoàn tất!")
		fmt.Println("🚀 SENT Agent đã được chuyển sang chế độ chạy ẩn (Daemon Mode).")
		fmt.Printf("📌 PID tiến trình ngầm: [%d]\n", cmd.Process.Pid)
		os.Exit(0)
	}

	utils.IgnoreSIGHUP()
}
