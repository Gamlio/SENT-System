package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net"
)

// SignPayload creates an HMAC-SHA256 signature for a given payload.
// This is used to authenticate requests from the agent to the backend.
func SignPayload(payload []byte, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// GetOutboundIP gets the preferred outbound ip of this machine.
// This is a helper function that is also used by the network collector.
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Printf("Warning: Could not determine outbound IP: %v", err)
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		log.Println("Warning: Type assertion for UDPAddr failed")
		return "127.0.0.1"
	}
	return localAddr.IP.String()
}
