package middleware

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	"github.com/gin-gonic/gin"
)

// RequirePermission kiểm tra quyền của người dùng dựa trên tên trường trong struct User
func RequirePermission(permissionField string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy user_id đã được lưu vào context bởi middleware AuthRequired
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Yêu cầu xác thực tài khoản"})
			c.Abort()
			return
		}

		var user models.User
		// Truy vấn thông tin user từ Postgres để lấy quyền mới nhất
		if err := database.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Người dùng không tồn tại hoặc đã bị xóa"})
			c.Abort()
			return
		}

		// Ánh xạ tên quyền truyền vào với các trường boolean trong struct User
		hasPermission := false
		switch permissionField {
		case "PermIncidentAction":
			hasPermission = user.PermIncidentAction
		case "PermIncidentView":
			hasPermission = user.PermIncidentView
		case "PermAssetAction":
			hasPermission = user.PermAssetAction
		case "PermAssetView":
			hasPermission = user.PermAssetView
		case "PermPolicyManage":
			hasPermission = user.PermPolicyManage
		case "PermApprovalFinal":
			hasPermission = user.PermApprovalFinal
		// Thêm các case khác tương ứng với các trường trong models.User nếu cần
		default:
			hasPermission = false
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Truy cập bị từ chối: Bạn không có quyền thực hiện hành động này",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
