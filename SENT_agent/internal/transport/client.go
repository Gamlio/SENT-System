package transport

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"sent_agent/internal/config"
	"sent_agent/internal/utils"

	"github.com/gorilla/websocket"
)

const (
	SERVER_URL = "http://localhost:8000/api/v1/agents/push"
	LOG_FILE   = "agent_history.log"
)

// XÓA bỏ trường CompanyCode ở đây
type Payload struct {
	Type     string      `json:"type"`
	LogType  string      `json:"log_type"`
	HWID     string      `json:"hwid"`
	Hostname string      `json:"hostname"`
	Data     interface{} `json:"data"`
}

type Client struct {
	LastHashes map[string]string
	LastStatus string
	Mutex      sync.Mutex
	WSConn     *websocket.Conn // Kết nối WebSocket duy nhất
}

var AgentClient = &Client{
	LastHashes: make(map[string]string),
	LastStatus: "PENDING",
}

// Hàm tạo chữ ký cho Agent
func generateSignature(payload []byte, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
func (c *Client) SendPayload(hwid, hostname, logType string, data interface{}, force bool) string {
	c.Mutex.Lock()
	currentHash := utils.CalculateHash(data)
	lastHash := c.LastHashes[logType]

	if !force && currentHash == lastHash {
		c.Mutex.Unlock()
		return "NO_CHANGE"
	}
	c.LastHashes[logType] = currentHash
	c.Mutex.Unlock()

	// Đã xóa config.Current.CompanyCode
	payload := Payload{
		Type:     "DATA",
		LogType:  logType,
		HWID:     hwid,
		Hostname: hostname,
		Data:     data,
	}

	jsonBytes, _ := json.Marshal(payload)
	signature := generateSignature(jsonBytes, config.Current.SecretKey)
	// Tạo Request mới để có thể nhét Header vào
	req, err := http.NewRequest("POST", SERVER_URL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "ERROR"
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sent-Signature", signature) // Đính kèm chữ ký

	// Thực hiện gửi
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)

	status := "OK"
	serverState := "ACTIVE"

	if err != nil {
		status = "FAIL: " + err.Error()
		serverState = "ERROR"
	} else {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		status = fmt.Sprintf("HTTP %d", resp.StatusCode)

		if resp.StatusCode == http.StatusOK {
			var respBody map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &respBody); err == nil {
				if state, ok := respBody["status"].(string); ok {
					serverState = state

					// Xóa sạch Hashes để gửi lại toàn bộ dữ liệu khi được duyệt
					if serverState == "ACTIVE" && c.LastStatus == "PENDING" {
						c.Mutex.Lock()
						c.LastHashes = make(map[string]string)
						c.Mutex.Unlock()
						fmt.Println("🚀 Máy đã được duyệt! Đang đồng bộ lại toàn bộ dữ liệu...")
					}
					c.LastStatus = serverState
				}
			}
		} else if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
			var errResp map[string]interface{}
			json.Unmarshal(bodyBytes, &errResp)
			if state, ok := errResp["status"].(string); ok {
				serverState = state
				c.LastStatus = serverState
			}
		}
	}

	logToFile("DATA", logType, status)
	return serverState
}

func logToFile(pType, lType, status string) {
	f, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		line := fmt.Sprintf("[%s] [%s] %s | %s\n", time.Now().Format(time.DateTime), pType, lType, status)
		f.WriteString(line)
		f.Close()
	}
}

type AlertData struct {
	AlertType string `json:"alert_type"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
}

func (c *Client) SendAlert(hwid, hostname, alertType, message, severity string) {
	alertPayload := Payload{
		Type:     "ALERT",
		LogType:  "alert",
		HWID:     hwid,
		Hostname: hostname,
		Data: AlertData{
			AlertType: alertType,
			Message:   message,
			Severity:  severity,
		},
	}

	jsonBytes, _ := json.Marshal(alertPayload)
	signature := generateSignature(jsonBytes, config.Current.SecretKey)

	req, _ := http.NewRequest("POST", SERVER_URL, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sent-Signature", signature)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)

	if err == nil {
		fmt.Printf("🚨 Đã gửi cảnh báo khẩn cấp [%s] về SOC Server!\n", alertType)
		resp.Body.Close()
	}
}

// [MỚI] SendBaseline: Xóa bộ nhớ đệm và gửi dữ liệu chuẩn Zero Trust
func (c *Client) SendBaseline(hwid, hostname, logType string, data interface{}) {
	c.Mutex.Lock()
	// Xóa dấu vết cũ của loại log này để ép Agent gửi lại bản full
	delete(c.LastHashes, logType)
	c.Mutex.Unlock()

	// Gửi kèm flag baseline để Backend xử lý vào bảng Baseline riêng
	c.SendPayload(hwid, hostname, logType+"_baseline", data, false)
}

// [MỚI] ListenForCommands: Lắng nghe lệnh từ Dashboard qua WebSocket
func (c *Client) StartHybridCommunication(hwid string, onCommand func(string, interface{})) {
	go func() {
		for {
			u := url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/ws"} // Cấu hình từ config.BackendURL
			q := u.Query()
			q.Set("token", config.Current.SecretKey) // Dùng SecretKey làm phương thức xác thực
			q.Set("hwid", hwid)
			u.RawQuery = q.Encode()

			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				time.Sleep(10 * time.Second) // Thử lại sau 10s nếu rớt mạng
				continue
			}

			c.WSConn = conn
			fmt.Println("✅ [WebSocket] Đã thiết lập kênh lệnh hai chiều với Server")

			for {
				var msg struct {
					Type string      `json:"type"`
					Data interface{} `json:"data"`
				}
				if err := conn.ReadJSON(&msg); err != nil {
					break // Ngắt kết nối để reconnect
				}
				// Xử lý lệnh (VD: TRIGGER_BASELINE, ISOLATE, RESTART)
				onCommand(msg.Type, msg.Data)
			}
			c.WSConn = nil
		}
	}()
}
