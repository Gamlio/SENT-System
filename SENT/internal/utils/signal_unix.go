//go:build !windows

package utils

import (
	"os"
	"os/signal"
	"syscall"
)

// IgnoreSIGHUP bẫy và bỏ qua tín hiệu tắt terminal SIGHUP trên Linux/Unix
func IgnoreSIGHUP() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	go func() {
		for {
			<-c // Đọc tín hiệu và hủy bỏ nó, không cho tiến trình bị kill
		}
	}()
}
