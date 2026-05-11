package handlers // Đảm bảo package là handlers để khớp với main.go

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	incidents "SENT_backend/service/incidents/internal/service"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

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

func CreateIncident(c *gin.Context) {
	var req struct {
		AlertID string `json:"alert_id"`
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

	err := database.DB.Model(&models.Incident{}).
		Where("id = ? AND org_id = ?", id, c.GetUint("org_id")).
		Update("status", req.Status).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi cập nhật trạng thái"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã cập nhật trạng thái"})
}

func AddIncidentComment(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"message": "Tính năng bình luận đang được xử lý qua Audit Log"})
}

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

func CloseIncident(c *gin.Context) {
	var req struct {
		IncidentID uint   `json:"incident_id" binding:"required"`
		Note       string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	svc := incidents.IncidentService{}
	audit := models.IncidentAudit{
		IncidentID: req.IncidentID,
		UserID:     ptrUint(c.GetUint("user_id")),
		Content:    req.Note,
	}

	err := svc.CloseIncidentWorkflow(c.Request.Context(), &audit, c.GetUint("org_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Hồ sơ đã được đóng"})
}

func UploadAudit(c *gin.Context) {
	incidentID, _ := strconv.Atoi(c.PostForm("incident_id"))
	note := c.PostForm("note")
	form, _ := c.MultipartForm()
	files := form.File["files"]

	var fileNames []string
	for _, file := range files {
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
		c.SaveUploadedFile(file, "./uploads/audits/"+filename)
		fileNames = append(fileNames, filename)
	}

	svc := incidents.IncidentService{}
	audit := models.IncidentAudit{
		IncidentID: uint(incidentID),
		UserID:     ptrUint(c.GetUint("user_id")),
		ActionType: "EVIDENCE",
		Content:    note,
		Images:     fileNames,
		CreatedAt:  time.Now(),
	}

	err := svc.CreateChainedAudit(c.Request.Context(), &audit, c.GetUint("org_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lưu log"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã lưu bằng chứng"})
}

func ptrUint(u uint) *uint { return &u }
