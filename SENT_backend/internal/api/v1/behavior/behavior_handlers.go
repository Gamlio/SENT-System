package behavior

import (
	"net/http"
	"sent_backend/internal/models"
	behaviorSvc "sent_backend/internal/service/Behavior"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetBehaviors: Lấy danh sách các hành vi vi phạm (Alerts) từ MongoDB
func GetBehaviors(c *gin.Context) {
	orgID := c.GetUint("org_id") // Lấy từ Middleware Auth
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	svc := behaviorSvc.BehaviorService{}

	// [FIX] Sử dụng biến total để trả về cho Frontend
	alerts, total, err := svc.GetBehaviors(c.Request.Context(), orgID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy dữ liệu"})
		return
	}

	// Trả về cấu trúc object thay vì mảng đơn thuần
	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": alerts,
		"page":  page,
		"limit": limit,
	})
}
func GetBehaviorDetail(c *gin.Context) {
	id := c.Param("id")
	// [SECURITY] Lấy org_id từ context để đảm bảo đúng phạm vi truy cập
	orgID := c.GetUint("org_id")

	svc := behaviorSvc.BehaviorService{}

	// Gọi service đã thêm để lấy dữ liệu từ MongoDB
	alert, err := svc.GetBehaviorDetail(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy chi tiết hành vi hoặc ID không hợp lệ"})
		return
	}

	// [SECURITY CHECK] Chốt chặn cuối cùng: Đảm bảo cảnh báo này thuộc về đúng tổ chức
	if alert.OrgID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền xem chi tiết hành vi này"})
		return
	}

	c.JSON(http.StatusOK, alert)
}
func CreateIncidentHandler(c *gin.Context) {
	var req struct {
		AlertID     string `json:"alert_id" binding:"required"`
		Severity    string `json:"severity" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu thiếu"})
		return
	}

	svc := behaviorSvc.BehaviorService{}
	incident, err := svc.CreateIncident(c.Request.Context(), req.AlertID, c.GetUint("org_id"), models.Incident{
		Severity:    req.Severity,
		Description: req.Description,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"incident_id": incident.ID})
}
