package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

// ============================= PHẦN 1: CUSTOM VALIDATORS =============================

// ValidatePasswordStrength: Kiểm tra độ mạnh của mật khẩu
func ValidatePasswordStrength(password string) (bool, string) {
	if len(password) < 8 {
		return false, "Mật khẩu phải có ít nhất 8 ký tự"
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasDigit   = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case !unicode.IsLetter(char) && !unicode.IsDigit(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return false, "Mật khẩu phải chứa ít nhất một ký tự in hoa"
	}
	if !hasLower {
		return false, "Mật khẩu phải chứa ít nhất một ký tự in thường"
	}
	if !hasDigit {
		return false, "Mật khẩu phải chứa ít nhất một chữ số"
	}
	if !hasSpecial {
		return false, "Mật khẩu phải chứa ít nhất một ký tự đặc biệt (!@#$%^&*)"
	}

	return true, ""
}

// ============================= PHẦN 2: MIDDLEWARE CHAIN =============================

// SecurityValidationMiddleware: Middleware chính để kiểm tra bảo mật toàn chuỗi request
func SecurityValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Kiểm tra Content-Type
		contentType := c.GetHeader("Content-Type")

		// Tách riêng logic: Upload tài liệu (lớn) vs JSON API (nhỏ)
		if strings.HasPrefix(contentType, "multipart/form-data") {
			// UC-02: Cho phép tối đa 10MB đối với upload tài liệu PDF/Word
			if c.Request.ContentLength > 10*1024*1024 {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Dung lượng file upload quá lớn (max 10MB)"})
				c.Abort()
				return
			}
		} else {
			// Các request JSON bình thường (Telemetry, Login...) vẫn bị khóa ở 100KB
			if c.Request.ContentLength > 100*1024 {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request body quá lớn (max 100KB)"})
				c.Abort()
				return
			}
			if !strings.Contains(contentType, "application/json") && c.Request.Method != http.MethodGet {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Content-Type không hợp lệ (yêu cầu application/json hoặc multipart/form-data)"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// assetAuthSecurityMiddleware: Kiểm tra security riêng biệt cho asset endpoints
func assetAuthSecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		hwid := c.Param("hwid")

		// Validate AssetHWID  format (max 64 eric)
		if len(hwid) > 64 || !regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(hwid) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AssetHWID  format"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateUserCreationMiddleware: Middleware riêng cho user creation endpoints
func ValidateUserCreationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Đọc body vào một buffer để có thể đọc lại sau
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)

		// Phân tích bằng struct cụ thể, KHÔNG dùng map[string]interface{} để tránh Reflection GC
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(tee).Decode(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			c.Abort()
			return
		}
		// Nạp lại dữ liệu vào Body để Handler phía sau có thể đọc
		c.Request.Body = io.NopCloser(&buf)

		// Validate password strength
		if req.Password != "" {
			if valid, msg := ValidatePasswordStrength(req.Password); !valid {
				c.JSON(http.StatusBadRequest, gin.H{"error": msg})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ValidateAssetPayloadMiddleware: Middleware riêng cho asset telemetry endpoints
func ValidateAssetPayloadMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Phân tích bằng struct cụ thể để tối ưu CPU và RAM
		var req struct {
			HWID    string `json:"hwid"`
			LogType string `json:"log_type"`
		}

		// Tận dụng raw_body từ AssetHMACAuth nếu có để tránh copy bộ nhớ lần 2
		if rawBody, exists := c.Get("raw_body"); exists {
			if err := json.Unmarshal(rawBody.([]byte), &req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset payload format"})
				c.Abort()
				return
			}
		} else {
			// Fallback nếu không có HMAC middleware phía trước
			var buf bytes.Buffer
			tee := io.TeeReader(c.Request.Body, &buf)
			if err := json.NewDecoder(tee).Decode(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset payload format"})
				c.Abort()
				return
			}
			c.Request.Body = io.NopCloser(&buf)
		}

		// Validate AssetHWID
		if req.HWID != "" {
			if len(req.HWID) > 64 || !regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(req.HWID) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AssetHWID  format"})
				c.Abort()
				return
			}
		}

		// Validate LogType
		if req.LogType != "" {
			if !regexp.MustCompile(`^[a-zA-Z0-9_]*$`).MatchString(req.LogType) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid LogType format"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
