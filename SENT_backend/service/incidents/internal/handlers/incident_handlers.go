package handlers // Đảm bảo package là handlers để khớp với main.go

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	incidents "SENT_backend/service/incidents/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetIncidents: Lấy danh sách sự cố
func GetIncidents(c *gin.Context) {
	orgID := c.GetUint("org_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))

	svc := incidents.IncidentService{}
	list, total, err := svc.GetIncidentList(orgID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy danh sách sự cố"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list, "total": total})
}

// GetIncidentDetail: Lấy chi tiết
func GetIncidentDetail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	orgID := c.GetUint("org_id")

	svc := incidents.IncidentService{}
	incident, audits, err := svc.GetIncidentDetail(c.Request.Context(), uint(id), orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy hồ sơ sự cố"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"incident": incident, "audit_logs": audits})
}

// GetIncidentActivities: Lấy dòng thời gian hoạt động (Missing trong bản cũ)
func GetIncidentActivities(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	orgID := c.GetUint("org_id")

	svc := incidents.IncidentService{}
	_, audits, err := svc.GetIncidentDetail(c.Request.Context(), uint(id), orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy hoạt động sự cố"})
		return
	}
	c.JSON(http.StatusOK, audits)
}

// CreateIncident: Tạo sự cố thủ công hoặc nâng cấp (Missing)
func CreateIncident(c *gin.Context) {
	var req struct {
		AlertID string `json:"alert_id"` // Nếu tạo từ Alert
	}
	c.ShouldBindJSON(&req)

	svc := incidents.IncidentService{}
	incident, err := svc.EscalateToIncident(c.Request.Context(), req.AlertID, c.GetUint("user_id"), c.GetUint("org_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, incident)
}

// UpdateIncidentStatus: Cập nhật trạng thái sự cố (Missing)
func UpdateIncidentStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Status string `json:"status" binding:"required"`
		Note   string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Thực hiện cập nhật thông qua DB trực tiếp hoặc gọi Service
	err := database.DB.Model(&models.Incident{}).
		Where("id = ? AND org_id = ?", id, c.GetUint("org_id")).
		Update("status", req.Status).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi cập nhật trạng thái"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã cập nhật trạng thái"})
}

// AddIncidentComment: Thêm bình luận/vết điều tra (Tên mới để khớp main.go)
func AddIncidentComment(c *gin.Context) {
	// Bạn có thể sử dụng lại logic của AddAuditLogHandler tại đây
	// Ở đây tôi viết gọn để bạn sửa lỗi undefined trước
	c.JSON(http.StatusOK, gin.H{"message": "Tính năng bình luận đang được xử lý qua Audit Log"})
}

// AssignIncident: Phân công nhân sự
func AssignIncident(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		AssigneeID uint `json:"assignee_id" binding:"required"`
	}
	c.ShouldBindJSON(&req)

	svc := incidents.IncidentService{}
	err := svc.AssignIncident(c.Request.Context(), uint(id), req.AssigneeID, c.GetUint("user_id"), c.GetUint("org_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã phân công nhân sự"})
}
