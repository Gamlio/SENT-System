package config

import (
	"bufio"
	"bytes"
	"encoding/base64"
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

		// Đảm bảo BackendURL luôn được cập nhật từ biến môi trường kể cả khi đã có file config cũ
		updateBackendURL()

		if Current.SecretKey != "" {
			Current.HWID = hwid
			// Giải mã SecretKey khi nạp vào RAM
			Current.SecretKey = deobfuscateKey(Current.SecretKey, hwid)
			return
		}
	}

	// Nếu chưa đăng ký, tiến hành Bootstrap
	updateBackendURL()

	fmt.Println("===========================================")
	fmt.Println("   🛡️ KÍCH HOẠT SENT SENSOR (CLI MODE)")
	fmt.Println("===========================================")
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

	// Mã hóa (Obfuscate) SecretKey trước khi lưu xuống đĩa
	rawSecret := result["secret_key"].(string)
	Current.SecretKey = obfuscateKey(rawSecret, hwid)

	// Lưu xuống đĩa cứng
	data, _ := json.MarshalIndent(Current, "", "  ")
	os.WriteFile(CONFIG_FILE, data, 0600)

	// Khôi phục lại SecretKey gốc trên RAM cho các thao tác tiếp theo
	Current.SecretKey = rawSecret

	fmt.Println("✅ Đăng ký thành công! Máy trạm đang chờ SOC phê duyệt (Status: PENDING).")
	time.Sleep(2 * time.Second)
	return true
}

func GetHWID() string {
	return Current.HWID
}

// GetPolicyURL returns the full URL to the policy sync endpoint on the server.
func GetPolicyURL() string {
	return Current.BackendURL + "/api/v1/policies"
}

// GetSecretKey returns the secret key used for signing requests.
func GetSecretKey() string {
	return Current.SecretKey
}

func updateBackendURL() {
	envURL := os.Getenv("SENT_BACKEND_URL")
	if envURL != "" {
		Current.BackendURL = envURL
	} else {
		// Fallback về domain của bạn nếu không tìm thấy trong .env
		Current.BackendURL = "https://sent.com.vn"
	}
}

// Mã hóa XOR cơ bản để tránh lưu SecretKey dưới dạng Plaintext
func obfuscateKey(secret, hwid string) string {
	b := []byte(secret)
	for i := range b {
		b[i] ^= hwid[i%len(hwid)]
	}
	return base64.StdEncoding.EncodeToString(b)
}

func deobfuscateKey(obfuscated, hwid string) string {
	b, _ := base64.StdEncoding.DecodeString(obfuscated)
	for i := range b {
		b[i] ^= hwid[i%len(hwid)]
	}
	return string(b)
}
