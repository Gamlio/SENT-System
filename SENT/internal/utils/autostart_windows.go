//go:build windows

package utils

import (
	"log"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// SetupAutoStart cấu hình ứng dụng khởi động cùng Windows
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

	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.ALL_ACCESS)
	if err != nil {
		log.Printf("⚠️ Không thể mở Registry Key HKLM...Run (Yêu cầu quyền Admin): %v", err)
		return
	}
	defer key.Close()

	err = key.SetStringValue("SENTAgent", `"`+exePath+`"`)
	if err != nil {
		log.Printf("⚠️ Không thể lưu Registry giá trị khởi động: %v", err)
	} else {
		log.Println("✅ Đã thiết lập khởi động cùng hệ thống trên Windows thành công.")
	}
}
