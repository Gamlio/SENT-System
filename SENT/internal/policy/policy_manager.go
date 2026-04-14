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
		if p.Category == "SOFTWARE" && p.PolicyType == "BLACKLIST" {
			blacklist = append(blacklist, strings.ToLower(p.Value))
		}
	}
	return blacklist
}

// StartSync chạy một goroutine để đồng bộ chính sách định kỳ.
func StartSync() {
	go func() {
		// Đồng bộ ngay lần đầu tiên
		fetchPolicies()

		ticker := time.NewTicker(5 * time.Minute) // Đồng bộ mỗi 5 phút
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
