package behavior

import (
	"net/http"
	behaviorSvc "sent_backend/internal/service/Behavior"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetBehaviors: Lấy danh sách các hành vi vi phạm (Alerts) từ MongoDB
func GetBehaviors(c *gin.Context) {
	orgID := c.GetUint("org_id")
	// [FIX] Thêm phân trang
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)

	svc := behaviorSvc.BehaviorService{}

	alerts, err := svc.GetBehaviors(c.Request.Context(), int64(orgID), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy danh sách hành vi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, alerts)
}
