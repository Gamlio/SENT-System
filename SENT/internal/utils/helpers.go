package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"sync/atomic"
)

var globalSequence int64

func GetGlobalSequence() int64 {
	return atomic.AddInt64(&globalSequence, 1)
}

func CalculateHash(data interface{}) string {
	bytes, _ := json.Marshal(data)
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

func MapToStruct(input interface{}, output interface{}) error {
	b, _ := json.Marshal(input)
	return json.Unmarshal(b, output)
}
func CalculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
