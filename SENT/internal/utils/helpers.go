package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
