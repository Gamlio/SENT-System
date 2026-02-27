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

		// Lấy org_id từ token (JWT lưu số dưới dạng float64)
		orgIDFloat, ok := claims["org_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token thiếu thông tin tổ chức"})
			c.Abort()
			return
		}
		orgID := uint(orgIDFloat)

		// 7. Truy vấn CHÍNH XÁC người dùng đó TẠI công ty đó
		var user models.User
		if err := database.DB.Where("username = ? AND org_id = ?", username, orgID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Người dùng không tồn tại"})
			c.Abort()
			return
		}

		// --- BẮT BUỘC PHẢI THÊM ĐOẠN NÀY ---
		// 8. Bơm thông tin vào Context để API CreateUser có thể lấy ra bằng c.Get("org_id")
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("org_id", *user.OrgID)
		c.Set("role", user.Role)

		// 9. Cấp phép cho request đi qua trạm kiểm soát để vào API chính
		c.Next()
	}
}
