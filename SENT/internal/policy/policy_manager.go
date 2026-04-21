package policy

import (
	"SENT/internal/config"
	"SENT/internal/utils"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
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
	policyCache []Policy
	mu          sync.RWMutex
)

// GetBlacklistedSoftware trả về danh sách tên các phần mềm bị cấm.
func GetBlacklistedSoftware() []string {
	mu.RLock()
	defer mu.RUnlock()

	var blacklist []string
	for _, p := range policyCache {
		if (strings.ToUpper(p.Category) == "SOFTWARE" || strings.ToUpper(p.Category) == "PROCESS") && p.PolicyType == "BLACKLIST" {
			blacklist = append(blacklist, strings.ToLower(p.Value))
		}
	}
	return blacklist
}

// CheckPolicy: Đánh giá tức thì luật cho một thực thể (VD: phần mềm, tiến trình) tại máy trạm
func CheckPolicy(category string, value string) (isViolation bool, reason string) {
	mu.RLock()
	defer mu.RUnlock()

	hasWhitelist := false
	inWhitelist := false

	for _, p := range policyCache {
		// Bỏ qua các rule không đúng danh mục
		if !strings.EqualFold(p.Category, category) {
			continue
		}

		match := strings.EqualFold(strings.TrimSpace(p.Value), strings.TrimSpace(value))

		// Phát hiện cấm (Blacklist)
		if p.PolicyType == "BLACKLIST" && match {
			return true, "Bị chặn bởi Blacklist"
		}
		// Kiểm tra có áp dụng Whitelist không (Zero Trust)
		if p.PolicyType == "WHITELIST" {
			hasWhitelist = true
			if match {
				inWhitelist = true
			}
		}
	}

	// Nếu hệ thống áp dụng Zero Trust cho danh mục này mà không có tên trong danh sách -> Chặn
	if hasWhitelist && !inWhitelist {
		return true, "Không nằm trong Whitelist (Zero Trust)"
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

	client := &http.Client{Timeout: 20 * time.Second}
	payload := map[string]string{"asset_hwid": config.GetHWID()}
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", config.GetPolicyURL(), bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Lỗi tạo request đồng bộ chính sách: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	signature := utils.SignPayload(jsonPayload, config.GetSecretKey())
	req.Header.Set("X-Sent-Signature", signature)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Đồng bộ chính sách thất bại. Lỗi: %v, Status: %s", err, resp.Status)
		return
	}
	defer resp.Body.Close()

	var newPolicies []Policy
	if err := json.NewDecoder(resp.Body).Decode(&newPolicies); err == nil {
		mu.Lock()
		policyCache = newPolicies
		mu.Unlock()
		log.Printf("Đồng bộ thành công %d chính sách.", len(newPolicies))
	}
}
