package handlers

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	policyService "SENT_backend/service/policies/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPolicies(c *gin.Context) {
	orgID := c.GetUint("org_id")
	category := c.Query("category")
	status := c.Query("status")

	svc := &policyService.PolicyService{}
	list := svc.GetPolicies(orgID, category, status)
	c.JSON(http.StatusOK, list)
}

func GetPolicyDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	orgID := c.GetUint("org_id")

	var policy models.Policy
	if err := database.DB.Where("id = ? AND org_id = ?", uint(id), orgID).First(&policy).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy chính sách"})
		return
	}
	c.JSON(http.StatusOK, policy)
}

func CreatePolicy(c *gin.Context) {
	var req models.Policy
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	orgID := c.GetUint("org_id")
	username, _ := c.Get("username")

	svc := &policyService.PolicyService{}
	err := svc.CreatePolicyRequest(req, orgID, username.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu phê duyệt chính sách."})
}

func UpdatePolicy(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	orgID := c.GetUint("org_id")

	var req models.Policy
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	err := database.DB.Model(&models.Policy{}).
		Where("id = ? AND org_id = ?", uint(id), orgID).
		Updates(map[string]interface{}{
			"title":           req.Title,
			"value":           req.Value,
			"approval_status": "PENDING",
			"is_active":       false,
		}).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi cập nhật"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã cập nhật và đang chờ phê duyệt lại"})
}

func DeletePolicy(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	orgID := c.GetUint("org_id")
	username, _ := c.Get("username")

	svc := &policyService.PolicyService{}
	if err := svc.DeletePolicyRequest(uint(id), orgID, username.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu xóa chính sách."})
}

func ApprovePolicy(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	orgID := c.GetUint("org_id")
	username, _ := c.Get("username")

	err := database.DB.Model(&models.Policy{}).
		Where("id = ? AND org_id = ?", uint(id), orgID).
		Updates(map[string]interface{}{
			"approval_status": "APPROVED",
			"is_active":       true,
			"approved_by":     username,
		}).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi phê duyệt"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Chính sách đã được kích hoạt"})
}
