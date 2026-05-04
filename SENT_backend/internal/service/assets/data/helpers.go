package data

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	//"log"
)

func CalculateHash(data interface{}) string {
	bytes, _ := json.Marshal(data)
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

func MapToStruct(input interface{}, output interface{}) error {
	b, _ := json.Marshal(input)
	return json.Unmarshal(b, output)
}

// GetBytesFromData tối ưu việc ép kiểu data từ Interface sang Byte Array
// Hỗ trợ cả cơ chế gửi lẻ (nhận json.RawMessage) và cơ chế gửi gom nhóm (nhận map/slice interface)
func GetBytesFromData(data interface{}) ([]byte, error) {
	switch v := data.(type) {
	case json.RawMessage:
		return v, nil
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		// Fallback an toàn cho trường hợp data là một object đã được Unmarshal từ Batching
		return json.Marshal(data)
	}
}
