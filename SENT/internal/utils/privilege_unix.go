//go:build !windows

package utils

import "os"

// IsAdmin checks if the process is running with root privileges on Unix-like systems.
func IsAdmin() bool {
	return os.Geteuid() == 0
}
