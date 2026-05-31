package data

import (
	"SENT_backend/pkg/models"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

var behaviorClient = &http.Client{
	Timeout: 5 * time.Second,
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

func GetBytesFromData(data interface{}) ([]byte, error) {
	switch v := data.(type) {
	case json.RawMessage:
		return v, nil
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return json.Marshal(data)
	}
}
func SendBehaviorLog(category string, value string, title string, desc string, priority string, asset models.Asset) {
	go func() {
		payload := map[string]interface{}{
			"asset": map[string]interface{}{
				"asset_hwid":     asset.AssetHWID,
				"hostname":       asset.Hostname,
				"ip_address":     asset.IPAddress,
				"org_id":         asset.OrgID,
				"department_tag": asset.DepartmentTag,
			},
			"category": category,
			"value":    value,
			"title":    title,
			"desc":     desc,
			"priority": priority,
		}

		jsonData, _ := json.Marshal(payload)
		url := "http://behavior-service:8000/api/v1/behaviors/log"

		resp, err := behaviorClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("⚠️ Lỗi kết nối sang Behavior Service: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("⚠️ Behavior Service từ chối log (%d) cho máy %s\n", resp.StatusCode, asset.AssetHWID)
		}
	}()
}
func SendToPolicyBaseline(orgID uint, hwid string, category string, values []map[string]interface{}) {
	if len(values) == 0 {
		return
	}

	go func() {
		payload := map[string]interface{}{
			"org_id":     orgID,
			"asset_hwid": hwid,
			"category":   category,
			"values":     values,
		}
		jsonData, _ := json.Marshal(payload)

		url := "http://policy-service:8000/api/v1/policies/internal/baseline"

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("⚠️ Không thể gửi Baseline sang Policy Service: %v\n", err)
			return
		}
		defer resp.Body.Close()
	}()
}
