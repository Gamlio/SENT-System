package middleware

import (
	"fmt"
	"net/http"
	"time"

	"sent_backend/internal/cache"

	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware giới hạn số lượng request
func RateLimitMiddleware(reqPerSec int, burst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("ratelimit:global:%s", ip)

		// Áp dụng Rate Limit Fixed Window qua Redis. Burst được coi như limit trong 1 giây.
		if !cache.AllowRequest(key, int64(burst), time.Second) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Tốc độ yêu cầu quá nhanh. Vui lòng thử lại sau.",
			})
			return
		}
		c.Next()
	}
}

// ============================= asset FLOODING PROTECTION =============================

// AssetFloodProtectionMiddleware: Middleware chặn asset flooding
func AssetFloodProtectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Thay vì parse toàn bộ body, ta đọc X-Asset-HWID từ Header để tránh thắt cổ chai GC
		hwid := c.GetHeader("X-Asset-HWID")

		if hwid == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing X-Asset-HWID header"})
			c.Abort()
			return
		}

		// Kiểm tra flood
		bodySize := c.Request.ContentLength
		if bodySize < 0 {
			bodySize = 0
		}

		allowed, reason := cache.RecordAssetFlood(hwid, bodySize)
		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": fmt.Sprintf("Cảnh báo từ Sentinel: %s", reason),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ============================= SPECIALIZED RATE LIMITERS =============================

// UserLoginRateLimitMiddleware: Rate limiter cho login (max 3 attempts/5 min)
func UserLoginRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("ratelimit:login:%s", ip)

		if !cache.AllowRequest(key, 3, 5*time.Minute) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many login attempts. Please try again after 5 minutes",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AssetEnrollRateLimitMiddleware: Rate limiter cho enrollment
func AssetEnrollRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("ratelimit:enroll:%s", ip)

		if !cache.AllowRequest(key, 5, 5*time.Second) { // Tối đa 5 request trong 5 giây
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

// AddIPToBlacklist: Thêm IP vào blacklist
func AddIPToBlacklist(ip string) {
	cache.SetBlacklist(ip, 24*time.Hour) // Block 24h qua Redis
}

// RemoveIPFromBlacklist: Xóa IP khỏi blacklist
func RemoveIPFromBlacklist(ip string) {
	// Xóa thông qua cache.RDB (Cần context)
	cache.RDB.Del(cache.Ctx, "blacklist:"+ip)
}

// IsIPBlacklisted: Kiểm tra IP có bị blacklist không
func IsIPBlacklisted(ip string) bool {
	return cache.IsBlacklisted(ip)
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
