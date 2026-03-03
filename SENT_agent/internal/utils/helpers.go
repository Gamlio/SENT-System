package utils // <--- BẮT BUỘC PHẢI LÀ UTILS

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net"
)

// CalculateHash: Viết hoa chữ đầu để Public
func CalculateHash(data interface{}) string {
	bytes, _ := json.Marshal(data)
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

func MapToStruct(input interface{}, output interface{}) error {
	b, _ := json.Marshal(input)
	return json.Unmarshal(b, output)
}

// GetOutboundIP: Viết hoa chữ đầu
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
