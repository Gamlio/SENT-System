package middleware

import (
	"fmt"
	"net/http"
	"os"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
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
			return []byte(os.Getenv("SECRET_KEY")), nil
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

		username := claims["sub"].(string)

		// Lấy org_id từ token (JWT lưu số dưới dạng float64)
		orgIDFloat, ok := claims["org_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token thiếu thông tin tổ chức"})
			c.Abort()
			return
		}
		orgID := uint(orgIDFloat)

		// 6. Truy vấn CHÍNH XÁC người dùng đó TẠI công ty đó
		var user models.User
		if err := database.DB.Where("username = ? AND org_id = ?", username, orgID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Người dùng không tồn tại"})
			c.Abort()
			return
		}

		// 7. Bơm thông tin vào Context (Chú ý các key: user_id, org_id)
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("org_id", *user.OrgID)
		c.Set("role", user.Role)

		c.Next()
	}
}
