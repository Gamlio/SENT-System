package incidents

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/incidents"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetIncidents: Lấy danh sách hỗ trợ phân trang
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
	c.JSON(http.StatusOK, gin.H{
		"items": list,
		"total": total,
	})
}

func GetIncidentDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	orgID := c.GetUint("org_id")

	svc := incidents.IncidentService{}
	incident, audits, err := svc.GetIncidentDetail(c.Request.Context(), uint(id), orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy hồ sơ sự cố"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"incident":   incident,
		"audit_logs": audits, // Luôn là mảng rỗng nếu chưa có log, không lo crash FE
	})
}

func AddAuditLogHandler(c *gin.Context) {
	// 1. Sử dụng c.PostForm để Gin tự động xử lý Multipart Form
	rawID := c.PostForm("incident_id")

	// Log chuỗi thô để kiểm tra nếu Frontend gửi sai
	fmt.Printf("DEBUG: Received raw incident_id from FE: '%s'\n", rawID)

	incidentID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || incidentID <= 0 {
		fmt.Printf("❌ LỖI: Không parse được ID. Lỗi: %v, Giá trị thô: '%s'\n", err, rawID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID sự cố không hợp lệ hoặc bị thiếu"})
		return
	}
	fmt.Printf("✅ DEBUG: Đang lưu Audit cho Incident ID: %d\n", incidentID)
	note := c.PostForm("note")
	orgID := c.GetUint("org_id")
	userID := int64(c.GetUint("user_id"))

	var user models.User
	if err := database.DB.Select("full_name").First(&user, userID).Error; err != nil {
		user.FullName = "Unknown User"
	}

	form, err := c.MultipartForm()

	var files []*multipart.FileHeader
	if err == nil && form != nil {
		files = form.File["files"]
	}

	var savedImages []string
	var imageHashes []string

	if files != nil {
		for _, file := range files {
			uploadDir := "./uploads/audits"
			if err := os.MkdirAll(uploadDir, 0755); err != nil {
				c.JSON(500, gin.H{"error": "Không thể tạo thư mục lưu bằng chứng"})
				return
			}
			filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
			savePath := filepath.Join(uploadDir, filename)

			if err := c.SaveUploadedFile(file, savePath); err == nil {
				savedImages = append(savedImages, filename)

				f, _ := file.Open()
				h := sha256.New()
				io.Copy(h, f)
				imageHashes = append(imageHashes, hex.EncodeToString(h.Sum(nil)))
				f.Close()
			}
		}
	}

	auditEntry := &models.IncidentAudit{
		IncidentID:  incidentID,
		UserID:      &userID,
		UserName:    user.FullName,
		ActionType:  "INVESTIGATE",
		Content:     note,
		Images:      savedImages,
		ImageHashes: imageHashes,
		IPAddress:   c.ClientIP(),
		CreatedAt:   time.Now(),
	}

	svc := incidents.IncidentService{}
	if err := svc.CreateChainedAudit(c.Request.Context(), auditEntry, orgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể niêm phong bằng chứng"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã lưu vết điều tra", "hash": auditEntry.AuditHash})
}

// CloseIncident: Kết thúc điều tra sự cố[cite: 58, 59]
func CloseIncident(c *gin.Context) {
	var req struct {
		IncidentID   uint   `json:"incident_id" binding:"required"`
		Note         string `json:"note" binding:"required"`
		EvidenceData string `json:"evidence_data" binding:"required"` // JSON Baseline scan
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đóng case không hợp lệ"})
		return
	}

	userID := int64(c.GetUint("user_id"))
	orgID := c.GetUint("org_id")

	var user models.User
	if err := database.DB.Select("full_name").First(&user, userID).Error; err != nil {
		user.FullName = "Unknown User"
	}

	auditEntry := &models.IncidentAudit{
		IncidentID:   int64(req.IncidentID),
		UserID:       &userID,
		UserName:     user.FullName,
		Content:      req.Note,
		EvidenceData: req.EvidenceData,
		IPAddress:    c.ClientIP(),
	}

	svc := incidents.IncidentService{}
	if err := svc.CloseIncidentWorkflow(c.Request.Context(), auditEntry, orgID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Hồ sơ đã được đóng và niêm phong", "hash": auditEntry.AuditHash})
}

// VerifyAuditIntegrity: Kiểm tra tính toàn vẹn chuỗi bằng chứng[cite: 58, 59]
func VerifyAuditIntegrity(c *gin.Context) {
	auditID := c.Param("audit_id")
	objID, _ := primitive.ObjectIDFromHex(auditID)

	var audit models.IncidentAudit
	if err := database.IncidentAuditCollection.FindOne(c.Request.Context(), bson.M{"_id": objID}).Decode(&audit); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy log"})
		return
	}

	svc := incidents.IncidentService{}
	isValid, msg, err := svc.VerifyIncidentAuditChain(c.Request.Context(), uint(audit.IncidentID), c.GetUint("org_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_intact": isValid, "message": msg})
}

// AssignIncident: Phân công điều tra viên[cite: 58, 59]
func AssignIncident(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		AssigneeID uint `json:"assignee_id" binding:"required"`
	}
	c.ShouldBindJSON(&req)

	svc := incidents.IncidentService{}
	if err := svc.AssignIncident(c.Request.Context(), uint(id), req.AssigneeID, c.GetUint("user_id"), c.GetUint("org_id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã phân công nhân sự"})
}
