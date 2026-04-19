package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1. ƯU TIÊN 1: Lấy từ Header (Dùng cho các API Axios bình thường)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// 2. ƯU TIÊN 2: Lấy từ Query URL (Dùng cho thẻ <img src="...?token=...">)
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		// 3. Nếu tìm cả 2 nơi đều không thấy -> Từ chối truy cập
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Yêu cầu mã xác thực (Token không được để trống)"})
			c.Abort()
			return
		}

		// 4. Giải mã và kiểm tra Token (Parse JWT)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("thuật toán không khớp: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token đã hết hạn hoặc không hợp lệ"})
			c.Abort()
			return
		}

		// 5. Trích xuất Claims từ Token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Không thể đọc dữ liệu Token"})
			c.Abort()
			return
		}

		username, ok := claims["sub"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token thiếu thông tin tài khoản (sub)"})
			c.Abort()
			return
		}

		// Lấy org_id từ token (JWT lưu số dưới dạng float64)
		orgIDFloat, ok := claims["org_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token thiếu thông tin tổ chức"})
			c.Abort()
			return
		}
		orgID := uint(orgIDFloat)

		// Lấy user_id từ token
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token thiếu thông tin ID người dùng (user_id)"})
			c.Abort()
			return
		}
		userID := uint(userIDFloat)

		// 6. Bơm thông tin vào Context (Bỏ qua truy vấn DB để tối ưu hiệu năng)
		c.Set("user_id", userID)
		c.Set("username", username)
		c.Set("org_id", orgID)

		c.Next()
	}
}
