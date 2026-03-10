package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"sent_agent/internal/config"
	"sent_agent/internal/utils"
)

const (
	SERVER_URL = "http://localhost:8000/api/v1/agents/push"
	LOG_FILE   = "agent_history.log"
)

type Payload struct {
	Type        string      `json:"type"`
	LogType     string      `json:"log_type"`
	HWID        string      `json:"hwid"`
	Hostname    string      `json:"hostname"`
	CompanyCode string      `json:"company_code"`
	Data        interface{} `json:"data"`
}

type Client struct {
	LastHashes map[string]string
	Mutex      sync.Mutex
}

var AgentClient = &Client{
	LastHashes: make(map[string]string),
}

// [CẬP NHẬT] Đổi hàm trả về string để main.go biết trạng thái (ACTIVE, PENDING, REJECTED)
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

	payload := Payload{
		Type:        "DATA",
		LogType:     logType,
		HWID:        hwid,
		Hostname:    hostname,
		CompanyCode: config.Current.CompanyCode,
		Data:        data,
	}

	jsonBytes, _ := json.Marshal(payload)
	resp, err := http.Post(SERVER_URL, "application/json", bytes.NewBuffer(jsonBytes))

	status := "OK"
	serverState := "ACTIVE" // Mặc định là Active nếu gọi thành công

	if err != nil {
		status = "FAIL: " + err.Error()
		serverState = "ERROR"
	} else {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		status = fmt.Sprintf("HTTP %d", resp.StatusCode)

		// [MỚI] Bắt lỗi 403 từ Server để lấy trạng thái PENDING/REJECTED
		if resp.StatusCode == http.StatusForbidden {
			var errResp map[string]interface{}
			json.Unmarshal(bodyBytes, &errResp)
			if state, ok := errResp["status"].(string); ok {
				serverState = state // Trả về "PENDING" hoặc "REJECTED"
			}
		}
	}

	logToFile("DATA", logType, status)
	return serverState // Trả trạng thái về cho main.go xử lý
}

func logToFile(pType, lType, status string) {
	f, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		line := fmt.Sprintf("[%s] [%s] %s | %s\n", time.Now().Format(time.DateTime), pType, lType, status)
		f.WriteString(line)
		f.Close()
	}
}

// Thêm cấu trúc cho Dữ liệu Cảnh báo
type AlertData struct {
	AlertType string `json:"alert_type"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
}

// Hàm gửi cảnh báo khẩn cấp (Bỏ qua kiểm tra Hash, gửi lập tức)
func (c *Client) SendAlert(hwid, hostname, alertType, message, severity string) {
	alertPayload := Payload{
		Type:        "ALERT", // Phân biệt với DATA bình thường
		LogType:     "alert",
		HWID:        hwid,
		Hostname:    hostname,
		CompanyCode: config.Current.CompanyCode,
		Data: AlertData{
			AlertType: alertType,
			Message:   message,
			Severity:  severity,
		},
	}

	jsonBytes, _ := json.Marshal(alertPayload)
	resp, err := http.Post(SERVER_URL, "application/json", bytes.NewBuffer(jsonBytes))

	if err == nil {
		fmt.Printf("🚨 Đã gửi cảnh báo khẩn cấp [%s] về SOC Server!\n", alertType)
		resp.Body.Close()
	}
}
