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

const CONFIG_FILE = "config/sent_config.json"

type Config struct {
	EnrollToken string `json:"enroll_token"`
	SecretKey   string `json:"secret_key"`
	BackendURL  string `json:"backend_url"`
	HWID        string `json:"hwid"`
}

var Current Config

func LoadOrBootstrap(hwid, hostname, ipAddress string) {
	if hwid == "" {
		fmt.Println("❌ Lỗi nghiêm trọng: Không thể định danh thiết bị (HWID rỗng). Vui lòng kiểm tra lại quyền truy cập hệ thống.")
		os.Exit(1)
	}

	file, err := os.ReadFile(CONFIG_FILE)
	if err == nil {
		json.Unmarshal(file, &Current)

		// Đảm bảo BackendURL luôn được cập nhật từ biến môi trường kể cả khi đã có file config cũ
		updateBackendURL()

		if Current.SecretKey != "" {
			// Cảnh báo nếu HWID trong file khác với HWID của hệ thống hiện tại
			if Current.HWID != "" && Current.HWID != hwid {
				fmt.Printf("⚠️ CẢNH BÁO: HWID đã thay đổi từ %s sang %s. Quá trình giải mã SecretKey có thể thất bại và gây lỗi 403!\n", Current.HWID, hwid)
			}

			needsUpdate := Current.HWID == ""
			Current.HWID = hwid

			// Giải mã SecretKey khi nạp vào RAM
			Current.SecretKey = deobfuscateKey(Current.SecretKey, hwid)

			// Nếu cấu hình trước đó thiếu HWID, ghi đè lại file để đồng bộ
			if needsUpdate {
				tempSecret := Current.SecretKey
				Current.SecretKey = obfuscateKey(tempSecret, hwid)
				data, _ := json.MarshalIndent(Current, "", "  ")
				os.WriteFile(CONFIG_FILE, data, 0600)
				Current.SecretKey = tempSecret
			}
			return
		}
	}

	// Nếu chưa đăng ký, tiến hành Bootstrap
	if os.Getenv("SENT_DAEMON") == "1" {
		fmt.Println("❌ Lỗi: Không tìm thấy file cấu hình trong chế độ Daemon. Vui lòng chạy Agent ở chế độ foreground trước để thiết lập.")
		os.Exit(1)
	}

	updateBackendURL()

	fmt.Println("===========================================")
	fmt.Println("   🛡️ KÍCH HOẠT SENT SENSOR (CLI MODE)")
	fmt.Println("===========================================")
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">> Nhập Mã Cài Đặt (Enrollment Token): ")
		token, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("\n❌ Lỗi đọc mã cài đặt: %v\n", err)
			os.Exit(1)
		}
		token = strings.TrimSpace(token)

		if enrollToServer(token, hwid, hostname, ipAddress) {
			break
		}
	}
	Current.HWID = hwid
}

func enrollToServer(token, hwid, hostname, ipAddress string) bool {
	if hwid == "" {
		fmt.Println("❌ Lỗi: Không thể đăng ký do HWID rỗng.")
		return false
	}

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

	// Sửa lỗi: Cập nhật HWID vào struct TRƯỚC khi ghi xuống đĩa để tránh ghi rỗng
	Current.HWID = hwid

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
