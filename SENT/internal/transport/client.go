package transport

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"SENT/internal/config"
	"SENT/internal/utils"
)

// AssetClient quản lý các kết nối mạng của SENT
type AssetClient struct {
	HTTPClient *http.Client
}

// GetAssetClient khởi tạo một HTTP client mới
func GetAssetClient() *AssetClient {
	return &AssetClient{
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// SendPayload gửi dữ liệu log thông thường (POST)
func (c *AssetClient) SendPayload(hwid, hostname, logType string, data interface{}) string {

	payload := map[string]interface{}{
		"type":       "DATA",
		"asset_hwid": hwid,
		"hostname":   hostname,
		"log_type":   logType,
		"data":       data,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Lỗi encode JSON payload: %v", err)
		return "ERROR"
	}

	req, err := http.NewRequest("POST", config.Current.BackendURL+"/api/v1/assets/push", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "ERROR"
	}

	req.Header.Set("Content-Type", "application/json")

	// Tạo chữ ký HMAC để xác thực với Backend
	signature := utils.SignPayload(jsonPayload, config.Current.SecretKey)
	req.Header.Set("X-Sent-Signature", signature)

	// Bổ sung header chống Replay Attack
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	seq := utils.GetGlobalSequence() // [BẢO MẬT] Dùng Sequence toàn cục để tránh xung đột với Policy
	sequence := strconv.FormatInt(time.Now().UnixNano()+seq, 10)
	req.Header.Set("X-Sent-Timestamp", timestamp)
	req.Header.Set("X-Sent-Sequence", sequence)
	req.Header.Set("X-Asset-HWID", hwid)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("Lỗi kết nối tới server (%s): %v", logType, err)
		return "ERROR"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // Đọc nội dung lỗi từ Backend
		log.Printf("🚨 Server trả về lỗi %d: %s", resp.StatusCode, string(body))
		if resp.StatusCode == http.StatusForbidden {
			return "ISOLATED" // Trả về tín hiệu máy đã bị backend cách ly
		}
		return "ERROR"
	}

	return "OK"
}
