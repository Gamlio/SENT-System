//go:build windows

package utils

import "golang.org/x/sys/windows"

// IsAdmin checks if the process is running with elevated (Administrator) privileges on Windows.
// This is crucial for sensors that need to access system-level information like firewall status or antivirus threats.
func IsAdmin() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}
