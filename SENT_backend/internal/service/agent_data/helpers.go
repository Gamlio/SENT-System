package agent_data

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

//func AnalyzeBehaviorAI(hwid string, dataType string, data interface{}) {
//log.Printf("🤖 AI đang phân tích dữ liệu %s của máy %s...", dataType, hwid)
// Logic AI sẽ được cắm vào đây sau này
//}
