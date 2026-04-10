package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter quản lý các limiter cho từng IP
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  sync.Mutex
	r   rate.Limit
	b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		r:   r,
		b:   b,
	}
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

// RateLimitMiddleware giới hạn số lượng request
func RateLimitMiddleware(r rate.Limit, b int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(r, b)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.GetLimiter(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Tốc độ yêu cầu quá nhanh. Vui lòng thử lại sau.",
			})
			return
		}
		c.Next()
	}
}

// ============================= asset FLOODING PROTECTION =============================

// assetFloodTracker: Theo dõi flooding từ asset
type assetFloodTracker struct {
	mu            sync.RWMutex
	assets        map[string]*assetRequestStats
	blockList     map[string]time.Time // AssetID -> BlockedUntil
	cleanupTicker *time.Ticker
}

// assetRequestStats: Thống kê request của mỗi asset
type assetRequestStats struct {
	RequestCount     int
	BytesSent        int64
	LastRequestTime  time.Time
	SuspiciousEvents int
	LastResetTime    time.Time
}

// NewassetFloodTracker: Khởi tạo tracker
func NewassetFloodTracker() *assetFloodTracker {
	tracker := &assetFloodTracker{
		assets:    make(map[string]*assetRequestStats),
		blockList: make(map[string]time.Time),
	}

	// Cleanup mỗi 5 phút
	tracker.cleanupTicker = time.NewTicker(5 * time.Minute)
	go func() {
		for range tracker.cleanupTicker.C {
			tracker.cleanup()
		}
	}()

	return tracker
}

// RecordRequest: Ghi nhận request từ asset
func (aft *assetFloodTracker) RecordRequest(hwid string, bodySize int64) bool {
	aft.mu.Lock()
	defer aft.mu.Unlock()

	now := time.Now()

	// Kiểm tra xem AssetID có bị block không
	if blockTime, exists := aft.blockList[hwid]; exists {
		if now.Before(blockTime) {
			return false // Vẫn bị block
		}
		delete(aft.blockList, hwid) // Bỏ block
	}

	// Lấy stats của asset
	stats, exists := aft.assets[hwid]
	if !exists {
		aft.assets[hwid] = &assetRequestStats{
			RequestCount:    1,
			BytesSent:       bodySize,
			LastRequestTime: now,
			LastResetTime:   now,
		}
		return true
	}

	// Reset nếu vượt quá 1 phút
	if now.Sub(stats.LastResetTime) > 1*time.Minute {
		stats.RequestCount = 0
		stats.BytesSent = 0
		stats.SuspiciousEvents = 0
		stats.LastResetTime = now
	}

	// Kiểm tra:
	// - Max 100 requests/phút
	// - Max 50MB/phút
	// - Max 10 request/giây (trong burst)
	stats.RequestCount++
	stats.BytesSent += bodySize
	stats.LastRequestTime = now

	// Detect flooding
	timeSinceLastRequest := now.Sub(stats.LastRequestTime)
	if timeSinceLastRequest < 10*time.Millisecond && stats.RequestCount > 5 {
		// asset gửi >5 request trong <10ms = flooding
		stats.SuspiciousEvents++
	}

	// Block nếu vượt quá giới hạn
	if stats.RequestCount > 100 {
		aft.blockList[hwid] = now.Add(10 * time.Minute)
		return false
	}

	if stats.BytesSent > 50*1024*1024 { // 50MB limit
		aft.blockList[hwid] = now.Add(10 * time.Minute)
		return false
	}

	if stats.SuspiciousEvents > 5 {
		aft.blockList[hwid] = now.Add(30 * time.Minute)
		return false
	}

	return true
}

// cleanup: Xóa asset records cũ
func (aft *assetFloodTracker) cleanup() {
	aft.mu.Lock()
	defer aft.mu.Unlock()

	now := time.Now()
	for hwid, stats := range aft.assets {
		// Xóa asset nếu không gửi request trong 1 giờ
		if now.Sub(stats.LastRequestTime) > 1*time.Hour {
			delete(aft.assets, hwid)
		}
	}

	// Xóa block thời gian hết
	for hwid, blockTime := range aft.blockList {
		if now.After(blockTime) {
			delete(aft.blockList, hwid)
		}
	}
}

// Stop: Dừng tracker
func (aft *assetFloodTracker) Stop() {
	aft.cleanupTicker.Stop()
}

// Global tracker instance
var assetFloodTrackerInstance = NewassetFloodTracker()

// AssetFloodProtectionMiddleware: Middleware chặn asset flooding
func AssetFloodProtectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy AssetID từ request
		var req struct {
			AssetID string `json:"asset_hwid"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing AssetID in request"})
			c.Abort()
			return
		}

		if req.AssetID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "AssetID cannot be empty"})
			c.Abort()
			return
		}

		// Kiểm tra flood
		bodySize := c.Request.ContentLength
		if bodySize < 0 {
			bodySize = 0
		}

		if !assetFloodTrackerInstance.RecordRequest(req.AssetID, bodySize) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": fmt.Sprintf("asset %s detected suspicious activity. Temporarily blocked", req.AssetID),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ============================= SPECIALIZED RATE LIMITERS =============================

// UserAuthRateLimiter: Rate limiting cho user login (strict)
var userAuthRateLimiter = NewIPRateLimiter(1, 3) // 1 request/sec, burst 3

// UserLoginRateLimitMiddleware: Rate limiter cho login (max 3 attempts/5 min)
func UserLoginRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !userAuthRateLimiter.GetLimiter(ip).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many login attempts. Please try again after 5 minutes",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// assetEnrollRateLimiter: Rate limiting cho enrollment (strict)
var assetEnrollRateLimiter = NewIPRateLimiter(0.2, 5) // 0.2 req/sec = 1 req/5sec, burst 5

// AssetEnrollRateLimitMiddleware: Rate limiter cho enrollment
func AssetEnrollRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !assetEnrollRateLimiter.GetLimiter(ip).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many enrollment attempts. Please try again after 5 minutes",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ============================= IP BLACKLIST SUPPORT =============================

var (
	ipBlacklistMutex sync.RWMutex
	ipBlacklist      = make(map[string]bool)
)

// AddIPToBlacklist: Thêm IP vào blacklist
func AddIPToBlacklist(ip string) {
	ipBlacklistMutex.Lock()
	defer ipBlacklistMutex.Unlock()
	ipBlacklist[ip] = true
}

// RemoveIPFromBlacklist: Xóa IP khỏi blacklist
func RemoveIPFromBlacklist(ip string) {
	ipBlacklistMutex.Lock()
	defer ipBlacklistMutex.Unlock()
	delete(ipBlacklist, ip)
}

// IsIPBlacklisted: Kiểm tra IP có bị blacklist không
func IsIPBlacklisted(ip string) bool {
	ipBlacklistMutex.RLock()
	defer ipBlacklistMutex.RUnlock()
	return ipBlacklist[ip]
}

// IPBlacklistMiddleware: Middleware kiểm tra blacklist
func IPBlacklistMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if IsIPBlacklisted(ip) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Your IP address is blocked",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
