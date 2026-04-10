package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

// ============================= PHẦN 1: CUSTOM VALIDATORS =============================

// ValidateSQLInjection: Kiểm tra SQL Injection patterns
func ValidateSQLInjection(input string) bool {
	// Danh sách các pattern nguy hiểm của SQL Injection
	sqlInjectionPatterns := []string{
		`(?i)union.*select`,
		`(?i)select.*from`,
		`(?i)insert.*into`,
		`(?i)delete.*from`,
		`(?i)drop.*table`,
		`(?i)update.*set`,
		`(?i)alter.*table`,
		`(?i)exec.*\(`,
		`(?i)execute.*\(`,
		`'?\s*or\s*'?=?'?`,
		`'?\s*and\s*'?=?'?`,
		`--\s*|#`,
		`;.*delete`,
		`;.*drop`,
	}

	for _, pattern := range sqlInjectionPatterns {
		regex, _ := regexp.Compile(pattern)
		if regex.MatchString(input) {
			return false
		}
	}
	return true
}

// ValidateXSSPayload: Kiểm tra XSS patterns
func ValidateXSSPayload(input string) bool {
	xssPatterns := []string{
		`(?i)<script[^>]*>.*?</script>`,
		`(?i)javascript:`,
		`(?i)on\w+\s*=`,
		`(?i)<iframe.*>`,
		`(?i)<object.*>`,
		`(?i)<embed.*>`,
		`(?i)<svg.*>`,
	}

	for _, pattern := range xssPatterns {
		regex, _ := regexp.Compile(pattern)
		if regex.MatchString(input) {
			return false
		}
	}
	return true
}

// ValidateCommandInjection: Kiểm tra Command Injection patterns
func ValidateCommandInjection(input string) bool {
	cmdPatterns := []string{
		`[;&|>']`, // Shell metacharacters
		`\$\(`,    // Command substitution
		`\${`,     // Variable expansion
	}

	for _, pattern := range cmdPatterns {
		regex, _ := regexp.Compile(pattern)
		if regex.MatchString(input) {
			return false
		}
	}
	return true
}

// ValidatePathTraversal: Kiểm tra Path Traversal patterns
func ValidatePathTraversal(input string) bool {
	if strings.Contains(input, "..") || strings.Contains(input, "~") {
		return false
	}
	return true
}

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
		if !strings.Contains(contentType, "application/json") && c.Request.Method != "GET" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Content-Type phải là application/json"})
			c.Abort()
			return
		}

		// Kiểm tra Body Size (99KB max - chặn request quá lớn -> DoS protection)
		if c.Request.ContentLength > 100*1024 {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request body quá lớn (max 100KB)"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// InputSanitizationMiddleware: Middleware để sanitize input từ các endpoint nhạy cảm
func InputSanitizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Phân tích các trường query & form
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if !ValidateSQLInjection(value) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": fmt.Sprintf("Query parameter '%s' chứa SQL injection pattern", key),
					})
					c.Abort()
					return
				}

				if !ValidateXSSPayload(value) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": fmt.Sprintf("Query parameter '%s' chứa XSS payload", key),
					})
					c.Abort()
					return
				}

				if !ValidatePathTraversal(value) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": fmt.Sprintf("Query parameter '%s' chứa path traversal", key),
					})
					c.Abort()
					return
				}
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
		// 1. Đọc toàn bộ nội dung Body dưới dạng mảng byte
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Không thể đọc dữ liệu"})
			c.Abort()
			return
		}

		// 2. [QUAN TRỌNG] Nạp lại dữ liệu vào Body để Handler phía sau (Register/Login) có thể đọc
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 3. Phân tích bản sao byte để validate
		var req map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			c.Abort()
			return
		}

		// Validate password strength
		passwordRaw, exists := req["password"]
		if exists {
			password, ok := passwordRaw.(string)
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be string"})
				c.Abort()
				return
			}

			if valid, msg := ValidatePasswordStrength(password); !valid {
				c.JSON(http.StatusBadRequest, gin.H{"error": msg})
				c.Abort()
				return
			}
		}

		// Validate username không chứa SQL injection
		usernameRaw, exists := req["username"]
		if exists {
			username, ok := usernameRaw.(string)
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Username must be string"})
				c.Abort()
				return
			}

			if !ValidateSQLInjection(username) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Username contains invalid characters"})
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
		var req map[string]interface{}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset payload format"})
			c.Abort()
			return
		}

		// Validate AssetHWID
		hwidRaw, exists := req["hwid"]
		if exists {
			hwid, ok := hwidRaw.(string)
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "AssetHWID  must be string"})
				c.Abort()
				return
			}

			if len(hwid) > 64 || !regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(hwid) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AssetHWID  format"})
				c.Abort()
				return
			}
		}

		// Validate LogType
		logTypeRaw, exists := req["log_type"]
		if exists {
			logType, ok := logTypeRaw.(string)
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "LogType must be string"})
				c.Abort()
				return
			}

			// LogType must be eric
			if !regexp.MustCompile(`^[a-zA-Z0-9_]*$`).MatchString(logType) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid LogType format"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
