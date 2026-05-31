package handlers

import (
	"SENT_backend/pkg/models"
	behaviorSvc "SENT_backend/service/behavior/internal/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetBehaviors: Lấy danh sách các hành vi vi phạm (Alerts) từ MongoDB
func GetBehaviors(c *gin.Context) {
	orgID := c.GetUint("org_id") // Lấy từ Middleware Auth
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
	search := c.Query("search")
	severity := c.DefaultQuery("severity", "all")
	status := c.DefaultQuery("status", "all")

	svc := behaviorSvc.BehaviorService{}

	// [FIX] Sử dụng biến total để trả về cho Frontend
	alerts, total, err := svc.GetBehaviors(c.Request.Context(), orgID, page, limit, search, severity, status)
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

func HandleLogBehavior(c *gin.Context) {
	var req struct {
		// Sử dụng DTO rút gọn để tránh Recursive Validation từ models.Asset
		Asset struct {
			AssetHWID     string `json:"asset_hwid" binding:"required"`
			Hostname      string `json:"hostname"`
			IPAddress     string `json:"ip_address"`
			OrgID         uint   `json:"org_id" binding:"required"`
			DepartmentTag string `json:"department_tag"`
		} `json:"asset" binding:"required"`
		Category string `json:"category" binding:"required"`
		Value    string `json:"value"`
		Title    string `json:"title" binding:"required"`
		Desc     string `json:"desc"`
		Priority string `json:"priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("❌ Lỗi giải mã log hành vi: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ", "detail": err.Error()})
		return
	}

	// If priority is not provided in the payload, default to P3
	basePriority := req.Priority
	if basePriority == "" {
		basePriority = "P3"
	}

	// Khởi tạo instance models.Asset và gán giá trị từ DTO
	// Việc này giúp khớp kiểu dữ liệu cho Service mà vẫn tránh được lỗi Validation 400 của các trường Inventory
	assetModel := models.Asset{
		AssetHWID:     req.Asset.AssetHWID,
		Hostname:      req.Asset.Hostname,
		IPAddress:     req.Asset.IPAddress,
		OrgID:         req.Asset.OrgID,
		DepartmentTag: req.Asset.DepartmentTag,
	}

	svc := behaviorSvc.BehaviorService{}
	alert, err := svc.LogBehavior(c.Request.Context(), assetModel, req.Category, req.Value, req.Title, req.Desc, basePriority)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alert)
}
