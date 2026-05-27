package policy

import (
	"SENT/internal/config"
	"SENT/internal/utils"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Policy đại diện cho một luật được kéo về từ server.
type Policy struct {
	Category   string `json:"category"`
	Value      string `json:"value"`
	PolicyType string `json:"policy_type"` // "BLACKLIST" hoặc "WHITELIST"
}

var (
	// Cấu trúc: [Category][Value] -> PolicyType
	// Ví dụ: optimizedCache["SOFTWARE"]["chrome.exe"] = "WHITELIST"
	optimizedCache map[string]map[string]string
	mu             sync.RWMutex
	currentVersion int64 // Lưu trữ version hiện hành của bộ luật
)

// [HIỆU NĂNG] Tái sử dụng HTTP Client với cấu hình Connection Pooling (Keep-Alive)
var httpClient = &http.Client{
	Timeout: 20 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:      10,
		IdleConnTimeout:   30 * time.Second,
		DisableKeepAlives: false,
	},
}

// GetBlacklistedSoftware trả về danh sách tên các phần mềm bị cấm.
func GetBlacklistedSoftware() []string {
	mu.RLock()
	defer mu.RUnlock()

	var blacklist []string
	for cat, values := range optimizedCache {
		if cat == "SOFTWARE" || cat == "PROCESS" {
			for val, pType := range values {
				if pType == "BLACKLIST" {
					blacklist = append(blacklist, val)
				}
			}
		}
	}
	return blacklist
}

// CheckPolicy: Đánh giá tức thì luật cho một thực thể (VD: phần mềm, tiến trình) tại máy trạm
func CheckPolicy(category string, value string) (isViolation bool, reason string) {
	mu.RLock()
	defer mu.RUnlock()

	// -------------------------------------------------------------------------
	// CHỐT CHẶN AN TOÀN (CIRCUIT BREAKER):
	// Nếu bộ nhớ đệm cache hoàn toàn trống rỗng (Ví dụ: Mới bật máy, chưa kết nối mạng
	// hoặc máy chủ đang sập nên chưa kéo được luật lần nào), lập tức cho qua (Mặc định không vi phạm).
	// Điều này ngăn Agent "tự sát" hoặc khóa cứng toàn bộ các phần mềm nền của hệ điều hành.
	// -------------------------------------------------------------------------
	if len(optimizedCache) == 0 {
		return false, ""
	}

	cat := strings.ToUpper(category)
	val := strings.ToLower(strings.TrimSpace(value))

	values, exists := optimizedCache[cat]
	if !exists {
		// Nếu danh mục này (Ví dụ: USB, PORT) hoàn toàn không cấu hình luật gì trên SOC, mặc định an toàn.
		return false, ""
	}

	// Thực hiện phép tra cứu bản đồ băm nhanh O(1)
	pType, found := values[val]
	if found {
		if pType == "BLACKLIST" {
			return true, "Thiết bị/Hành vi nằm trong danh sách cấm của tổ chức (Blacklist)"
		}
		// Nếu tìm thấy trong luật Whitelist (Bao gồm cả cấu hình Baseline máy sạch đã đồng bộ về) -> Hợp lệ!
		return false, ""
	}

	// Kiểm tra xem danh mục (Category) này có áp dụng cơ chế Zero Trust (Có cấu hình Whitelist) hay không
	hasWhitelist := false
	for _, t := range values {
		if t == "WHITELIST" {
			hasWhitelist = true
			break
		}
	}

	// Nếu SOC có cấu hình danh sách trắng (Whitelist) cho danh mục này, mà phần mềm/cổng mạng hiện tại
	// không tìm thấy trong danh sách (found == false), chứng tỏ đây là một hành vi lạ/bất thường.
	if hasWhitelist {
		return true, "Hành vi bị từ chối do không đăng ký trong danh sách trắng (Zero Trust Violation)"
	}

	return false, ""
}

// StartSync chạy một goroutine để đồng bộ chính sách định kỳ.
func StartSync() {
	go func() {
		// Đồng bộ ngay lần đầu tiên
		fetchPolicies()

		// Tăng tần suất đồng bộ lên 1 phút để phản ứng nhanh hơn với các lệnh chặn từ SOC
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			fetchPolicies()
		}
	}()
}

// fetchPolicies là hàm thực hiện gọi API về backend.
func fetchPolicies() {
	log.Println("Đang đồng bộ chính sách bảo vệ từ server...")

	payload := map[string]interface{}{
		"asset_hwid": config.GetHWID(),
		"version":    currentVersion, // Gửi kèm version để server đối chiếu
	}
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", config.GetPolicyURL(), bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Lỗi tạo request đồng bộ chính sách: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	signature := utils.SignPayload(jsonPayload, config.GetSecretKey())
	req.Header.Set("X-Sent-Signature", signature)

	// [BẢO MẬT] Bổ sung đầy đủ các Header chống Replay Attack để đồng bộ với Middleware Backend
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	seq := utils.GetGlobalSequence()
	sequence := strconv.FormatInt(time.Now().UnixNano()+seq, 10)
	req.Header.Set("X-Sent-Timestamp", timestamp)
	req.Header.Set("X-Sent-Sequence", sequence)
	req.Header.Set("X-Asset-HWID", config.GetHWID())

	resp, err := httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Đồng bộ chính sách thất bại. Lỗi: %v, Status: %s", err, resp.Status)
		return
	}
	defer resp.Body.Close()

	var respData struct {
		Status   string   `json:"status"`
		Version  int64    `json:"version"`
		Policies []Policy `json:"policies"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respData); err == nil {
		if respData.Status == "not_modified" {
			// Bỏ qua nếu luật không có gì thay đổi
			return
		}

		// Build cache mới O(1)
		newCache := make(map[string]map[string]string)
		for _, p := range respData.Policies {
			cat := strings.ToUpper(p.Category)
			val := strings.ToLower(p.Value)
			if newCache[cat] == nil {
				newCache[cat] = make(map[string]string)
			}
			newCache[cat][val] = p.PolicyType
		}
		mu.Lock()
		optimizedCache = newCache
		currentVersion = respData.Version
		mu.Unlock()
		log.Printf("Đồng bộ thành công %d chính sách. (Version: %d)", len(respData.Policies), currentVersion)
	}
}
