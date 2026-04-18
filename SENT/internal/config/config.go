package config

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const CONFIG_FILE = "sent_config.json"

type Config struct {
	EnrollToken string `json:"enroll_token"`
	SecretKey   string `json:"secret_key"`
	BackendURL  string `json:"backend_url"`
	HWID        string `json:"hwid"`
}

var Current Config

func LoadOrBootstrap(hwid, hostname, ipAddress string) {
	file, err := os.ReadFile(CONFIG_FILE)
	if err == nil {
		json.Unmarshal(file, &Current)
		// Nếu đã có SecretKey tức là đăng ký thành công rồi
		if Current.SecretKey != "" {
			Current.HWID = hwid // Cập nhật HWID vào bộ nhớ để các Sensor sử dụng
			return
		}
	}

	// Nếu chưa đăng ký, tiến hành Bootstrap
	fmt.Println("===========================================")
	fmt.Println("   🛡️ KÍCH HOẠT SENT SENSOR (CLI MODE)")
	fmt.Println("===========================================")
	Current.BackendURL = os.Getenv("BackendURL")
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">> Nhập Mã Cài Đặt (Enrollment Token): ")
		token, _ := reader.ReadString('\n')
		token = strings.TrimSpace(token)

		if enrollToServer(token, hwid, hostname, ipAddress) {
			break
		}
	}
	Current.HWID = hwid
}

func enrollToServer(token, hwid, hostname, ipAddress string) bool {
	payload := map[string]string{
		"asset_hwid": hwid,
		"hostname":   hostname,
		"ip_address": ipAddress,
		"token":      token,
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(Current.BackendURL+"/api/v1/assets/enroll", "application/json", bytes.NewBuffer(jsonData))

	if err != nil || resp.StatusCode != 200 {
		fmt.Println("❌ Lỗi: Mã cài đặt không đúng, đã hết hạn hoặc server không phản hồi!")
		return false
	}
	defer resp.Body.Close()

	// Đọc SecretKey do Backend trả về
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	Current.EnrollToken = token
	Current.SecretKey = result["secret_key"].(string)

	// Lưu xuống đĩa cứng
	data, _ := json.MarshalIndent(Current, "", "  ")
	os.WriteFile(CONFIG_FILE, data, 0644)

	fmt.Println("✅ Đăng ký thành công! Máy trạm đang chờ SOC phê duyệt (Status: PENDING).")
	time.Sleep(2 * time.Second)
	return true
}

// GetHWID returns the agent's hardware ID.
func GetHWID() string {
	return Current.HWID
}

// GetPolicyURL returns the full URL to the policy sync endpoint on the server.
func GetPolicyURL() string {
	return Current.BackendURL + "/api/v1/agent/policies"
}

// GetSecretKey returns the secret key used for signing requests.
func GetSecretKey() string {
	return Current.SecretKey
}
