package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net"
)

// SignPayload creates an HMAC-SHA256 signature for a given payload.
func SignPayload(payload []byte, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// GetOutboundIP gets the preferred outbound ip of this machine.
// This is a helper function that is also used by the network collector.
func GetOutboundIP() string {
	// Cách 1: Dial tới dải IP Private cục bộ thay vì 8.8.8.8
	// OS sẽ tự động đối chiếu Routing Table nội bộ để chọn Interface hợp lý mà không cần gửi gói tin ra Internet
	conn, err := net.Dial("udp", "10.255.255.255:80")
	if err == nil {
		defer conn.Close()
		if localAddr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			return localAddr.IP.String()
		}
	}

	// Cách 2: Fallback (Dự phòng) duyệt qua các card mạng nội bộ khi máy mất hoàn toàn Default Route
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Warning: Could not get network interfaces: %v", err)
		return "127.0.0.1"
	}

	for _, iface := range interfaces {
		// Bỏ qua loopback và các interface đang bị tắt
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// Ưu tiên IP v4 hợp lệ, bỏ qua link-local (169.254.x.x)
			if ip != nil && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && ip.To4() != nil {
				return ip.String()
			}
		}
	}
	return "127.0.0.1"
}
