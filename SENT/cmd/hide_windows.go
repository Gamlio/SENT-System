//go:build windows

package main

import (
	"golang.org/x/sys/windows"
)

// hideConsoleWindow sử dụng cơ chế bảo mật LazyDLL để ẩn cửa sổ console trên Windows
func hideConsoleWindow() {
	// Sử dụng NewLazySystemDLL để ép buộc Windows chỉ load từ thư mục System32, chống DLL Preloading
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	user32 := windows.NewLazySystemDLL("user32.dll")

	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	showWindow := user32.NewProc("ShowWindow")

	// Kiểm tra xem các hàm API có tồn tại trong hệ thống không
	if getConsoleWindow.Find() != nil || showWindow.Find() != nil {
		return
	}

	hwnd, _, _ := getConsoleWindow.Call()
	if hwnd != 0 {
		// 0 = SW_HIDE (Ẩn cửa sổ hoàn toàn dưới background)
		showWindow.Call(hwnd, 0)
	}
}
