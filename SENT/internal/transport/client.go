package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"SENT/internal/config"
	"SENT/internal/utils"

	"github.com/gorilla/websocket"
)

// AssetClient quản lý các kết nối mạng của Agent
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
func (c *AssetClient) SendPayload(hwid, hostname, logType string, data interface{}, isBaseline bool) string {
	if isBaseline {
		logType = logType + "_baseline"
	}

	payload := map[string]interface{}{
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

	req, err := http.NewRequest("POST", config.Current.BackendURL+"/api/v1/assets/data", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "ERROR"
	}

	req.Header.Set("Content-Type", "application/json")

	// Tạo chữ ký HMAC để xác thực với Backend
	signature := utils.SignPayload(jsonPayload, config.Current.SecretKey)
	req.Header.Set("X-Sent-Signature", signature)

	// Bổ sung header chống Replay Attack
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	sequence := strconv.FormatInt(time.Now().UnixNano(), 10)
	req.Header.Set("X-Sent-Timestamp", timestamp)
	req.Header.Set("X-Sent-Sequence", sequence)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("Lỗi kết nối tới server (%s): %v", logType, err)
		return "ERROR"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return "ISOLATED" // Trả về tín hiệu máy đã bị backend cách ly
	}

	return "OK"
}

// SendBaseline gửi dữ liệu cấu hình chuẩn ban đầu
func (c *AssetClient) SendBaseline(hwid, hostname, logType string, data interface{}) {
	c.SendPayload(hwid, hostname, logType, data, true)
}

// StartHybridCommunication thiết lập kết nối WebSocket để nhận lệnh realtime
func (c *AssetClient) StartHybridCommunication(hwid string, commandHandler func(string, interface{})) {
	go func() {
		for {
			// Tự động chuyển đổi HTTP/HTTPS sang WS/WSS
			wsURL := strings.Replace(config.Current.BackendURL, "http://", "ws://", 1)
			wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
			wsURL = fmt.Sprintf("%s/api/v1/ws?hwid=%s", wsURL, hwid)

			// Bổ sung bảo mật cho WebSocket
			headers := http.Header{}
			timestamp := strconv.FormatInt(time.Now().Unix(), 10)
			headers.Set("X-Sent-Timestamp", timestamp)
			headers.Set("X-Sent-Sequence", strconv.FormatInt(time.Now().UnixNano(), 10))
			headers.Set("X-Sent-Signature", utils.SignPayload([]byte(hwid+timestamp), config.Current.SecretKey))

			conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
			if err != nil {
				time.Sleep(5 * time.Second) // Thử lại sau 5s nếu không kết nối được
				continue
			}

			log.Println("📡 Đã kết nối WebSocket thành công tới Server!")

			for {
				var msg struct {
					Type      string      `json:"type"`
					Data      interface{} `json:"data"`
					Signature string      `json:"signature"`
				}

				err := conn.ReadJSON(&msg)
				if err != nil {
					log.Println("⚠️ Mất kết nối WebSocket, đang kết nối lại...")
					conn.Close()
					break
				}

				// Kiểm tra tính toàn vẹn của phản hồi (Chống MitM)
				if msg.Signature != "" {
					dataBytes, _ := json.Marshal(msg.Data)
					expectedSig := utils.SignPayload(dataBytes, config.Current.SecretKey)
					if msg.Signature != expectedSig {
						log.Println("🚨 CẢNH BÁO BẢO MẬT: Phát hiện lệnh điều khiển có chữ ký không hợp lệ. Đã từ chối thực thi!")
						continue
					}
				}

				// Chuyển lệnh nhận được ngược lại cho main.go xử lý
				commandHandler(msg.Type, msg.Data)
			}
		}
	}()
}
