package incidents

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/incidents"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetIncidents: Lấy danh sách
func GetIncidents(c *gin.Context) {
	orgID := c.GetUint("org_id")
	svc := incidents.IncidentService{}

	list, err := svc.GetIncidentList(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy danh sách sự cố"})
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetIncidentDetail: Chi tiết + Audit Logs
func GetIncidentDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	svc := incidents.IncidentService{}
	incident, audits, err := svc.GetIncidentDetail(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy sự cố"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"incident": incident, "audit_logs": audits})
}

// AssignIncident: Admin phân công
func AssignIncident(c *gin.Context) {
	idStr := c.Param("id")
	incidentID, _ := strconv.Atoi(idStr)

	var req struct {
		AssigneeID uint `json:"assignee_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Thiếu AssigneeID"})
		return
	}

	adminID := c.GetUint("user_id")
	svc := incidents.IncidentService{}
	if err := svc.AssignIncident(c.Request.Context(), uint(incidentID), req.AssigneeID, adminID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Đã phân công"})
}

// CloseIncident: Đóng sự cố
func CloseIncident(c *gin.Context) {
	var req struct {
		IncidentID   uint   `json:"incident_id" binding:"required"`
		Note         string `json:"note" binding:"required"`
		EvidenceData string `json:"evidence_data" binding:"required"`
		Images       string `json:"images"` // Thêm trường images để nhận link file bằng chứng
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Thiếu bằng chứng đóng case"})
		return
	}

	userID := c.GetUint("user_id")
	svc := incidents.IncidentService{}

	// FIX: Ép kiểu uint sang int64 để khớp với model MongoDB
	mongoIncidentID := int64(req.IncidentID)
	mongoUserID := int64(userID)

	auditEntry := &models.IncidentAudit{
		IncidentID:   mongoIncidentID, // Gán int64
		UserID:       &mongoUserID,    // Gán *int64
		Content:      req.Note,
		EvidenceData: req.EvidenceData,
		Images:       req.Images, // Gán giá trị images
		IPAddress:    c.ClientIP(),
	}

	if err := svc.CloseIncidentWorkflow(c.Request.Context(), auditEntry); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Đã đóng sự cố", "hash": auditEntry.AuditHash})
}

// VerifyAuditIntegrity: Kiểm tra tính toàn vẹn của toàn bộ chuỗi Log cho một sự cố.
// Nó lấy một audit_id bất kỳ, tìm ra sự cố cha, và xác minh toàn bộ chuỗi.
func VerifyAuditIntegrity(c *gin.Context) {
	auditID := c.Param("audit_id")
	objID, err := primitive.ObjectIDFromHex(auditID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID của audit log không hợp lệ"})
		return
	}

	// 1. Tìm bản ghi audit để lấy incident_id
	var audit models.IncidentAudit
	err = database.IncidentAuditCollection.FindOne(c.Request.Context(), bson.M{"_id": objID}).Decode(&audit)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Log không tồn tại"})
		return
	}

	// 2. Gọi service để kiểm tra toàn bộ chuỗi hash của sự cố liên quan
	svc := incidents.IncidentService{}
	isValid, message, err := svc.VerifyIncidentAuditChain(c.Request.Context(), uint(audit.IncidentID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi trong quá trình kiểm tra: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_intact": isValid,
		"message":   message,
	})
}

// UpdatePlaybookProgress:
func UpdatePlaybookProgress(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var req struct {
		Progress string `json:"progress"`
	}
	c.ShouldBindJSON(&req)
	database.DB.Model(&models.Incident{}).Where("id = ?", id).Update("playbook_progress", req.Progress)
	c.JSON(200, gin.H{"message": "Updated"})
}

// GetIncidentImage:
func GetIncidentImage(c *gin.Context) {
	filename := c.Param("filename")
	c.File("./uploads/incidents/" + filename)
}
