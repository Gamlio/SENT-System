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

	"sent_agent/internal/utils"
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
	Mutex      sync.Mutex
}

var AgentClient = &Client{
	LastHashes: make(map[string]string),
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
	resp, err := http.Post(SERVER_URL, "application/json", bytes.NewBuffer(jsonBytes))

	status := "OK"
	serverState := "ACTIVE"

	if err != nil {
		status = "FAIL: " + err.Error()
		serverState = "ERROR"
	} else {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		status = fmt.Sprintf("HTTP %d", resp.StatusCode)

		if resp.StatusCode == http.StatusForbidden {
			var errResp map[string]interface{}
			json.Unmarshal(bodyBytes, &errResp)
			if state, ok := errResp["status"].(string); ok {
				serverState = state
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
	// Đã xóa config.Current.CompanyCode
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
	resp, err := http.Post(SERVER_URL, "application/json", bytes.NewBuffer(jsonBytes))

	if err == nil {
		fmt.Printf("🚨 Đã gửi cảnh báo khẩn cấp [%s] về SOC Server!\n", alertType)
		resp.Body.Close()
	}
}
