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
		// 1. Lấy chuỗi Authorization từ Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Yêu cầu mã xác thực (Token)"})
			c.Abort()
			return
		}

		// 2. Tách chuỗi để lấy Token (Định dạng: Bearer <token>)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Định dạng Token không hợp lệ"})
			c.Abort()
			return
		}
		tokenString := parts[1] // Bây giờ tokenString đã được sử dụng

		// 3. Giải mã và kiểm tra Token (Parse JWT)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Kiểm tra thuật toán mã hóa (HS256)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("thuật toán không khớp: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("SECRET_KEY")), nil // Sử dụng SECRET_KEY từ .env
		})

		// 4. Xử lý lỗi Token hết hạn hoặc sai
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

		// 6. Truy vấn người dùng từ Database dựa trên "sub" (Username)
		username := claims["sub"].(string)
		var user models.User
		if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Người dùng không tồn tại"})
			c.Abort()
			return
		}

		// 7. Lưu thông tin User và OrgID vào Context để các API sau sử dụng
		c.Set("user", user)
		if user.OrgID != nil {
			c.Set("org_id", *user.OrgID) // Lưu OrgID để cô lập dữ liệu SME
		}

		c.Next() // Cho phép đi tiếp vào API chính
	}
}
