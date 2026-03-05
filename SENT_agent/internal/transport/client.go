package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	// [SỬA LỖI TẠI ĐÂY] Dùng đường dẫn tuyệt đối theo tên module
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

// Client quản lý trạng thái gửi
type Client struct {
	LastHashes map[string]string
	Mutex      sync.Mutex
}

// Tạo Client mới
var AgentClient = &Client{
	LastHashes: make(map[string]string),
}

func (c *Client) SendPayload(hwid, hostname, logType string, data interface{}, force bool) {
	c.Mutex.Lock()
	// Gọi hàm từ package utils (đã viết hoa)
	currentHash := utils.CalculateHash(data)
	oldHash := c.LastHashes[logType]
	packetType := "HEARTBEAT"
	var dataToSend interface{} = nil

	if force || currentHash != oldHash {
		packetType = "DATA"
		dataToSend = data
		c.LastHashes[logType] = currentHash
		fmt.Printf("⚡ [%s] Gửi dữ liệu mới...\n", logType)
	}
	c.Mutex.Unlock()

	if packetType == "HEARTBEAT" {
		dataToSend = nil
	} else if logType == "telemetry" {
		dataToSend = data
	}

	payload := Payload{
		Type:        packetType,
		LogType:     logType,
		HWID:        hwid,
		Hostname:    hostname,
		CompanyCode: config.Current.CompanyCode, // Gọi biến từ package config
		Data:        dataToSend,
	}

	jsonBytes, _ := json.Marshal(payload)
	resp, err := http.Post(SERVER_URL, "application/json", bytes.NewBuffer(jsonBytes))

	status := "OK"
	if err != nil {
		status = "FAIL: " + err.Error()
	} else {
		status = fmt.Sprintf("HTTP %d", resp.StatusCode)
		resp.Body.Close()
	}

	logToFile(packetType, logType, status)
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
