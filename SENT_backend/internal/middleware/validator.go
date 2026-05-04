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
			// Các request JSON bình thường (Telemetry, Login...) vẫn bị khóa ở 15MB
			if c.Request.ContentLength > 15*1024*1024 {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request body quá lớn (max 15MB)"})
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

		// Nới lỏng: Cho phép Chữ, Số, Gạch ngang (-), Gạch dưới (_), Hai chấm (:), và Dấu chấm (.)
		if len(hwid) > 128 || !regexp.MustCompile(`^[a-zA-Z0-9\-_:\.\{\}]*$`).MatchString(hwid) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Định dạng ID máy trạm không hợp lệ"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateUserCreationMiddleware: Middleware riêng cho user creation endpoints
func ValidateUserCreationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Không thể đọc dữ liệu request"})
			c.Abort()
			return
		}

		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			c.Abort()
			return
		}
		// Nạp lại dữ liệu vào Body để Handler phía sau có thể đọc
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

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
			HWID    string `json:"asset_hwid"`
			LogType string `json:"log_type"`
			Type    string `json:"type"`
		}

		// Tận dụng raw_body từ AssetHMACAuth nếu có để tránh copy bộ nhớ lần 2
		if rawBody, exists := c.Get("raw_body"); exists {
			// Debug: In ra dữ liệu thực tế Agent đang gửi lên TRƯỚC KHI unmarshal
			fmt.Println("Payload Received (Context):", string(rawBody.([]byte)))
			if err := json.Unmarshal(rawBody.([]byte), &req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset payload format"})
				c.Abort()
				return
			}
		} else {
			// Fallback nếu không có HMAC middleware phía trước
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Không thể đọc dữ liệu payload"})
				c.Abort()
				return
			}
			// Debug: In ra dữ liệu thực tế Agent đang gửi lên
			fmt.Println("Payload Received (Fallback):", string(bodyBytes))
			if err := json.Unmarshal(bodyBytes, &req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset payload format"})
				c.Abort()
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Validate AssetHWID
		if req.HWID != "" {
			// Tăng độ dài lên 128 ký tự để bao quát mọi loại thiết bị
			if len(req.HWID) > 128 || !regexp.MustCompile(`^[a-zA-Z0-9\-_:\.\{\}]*$`).MatchString(req.HWID) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Định dạng ID máy trạm không hợp lệ"})
				c.Abort()
				return
			}
		}

		// Validate LogType
		if req.LogType != "" {
			if !regexp.MustCompile(`^[a-zA-Z0-9\-_:\.]*$`).MatchString(req.LogType) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Định dạng LogType không hợp lệ"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
