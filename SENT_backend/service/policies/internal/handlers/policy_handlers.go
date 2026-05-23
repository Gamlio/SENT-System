package handlers

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	policyService "SENT_backend/service/policies/internal/service"
	"fmt"
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
	var req struct {
		Title      string   `json:"title"`
		Category   string   `json:"category"`
		PolicyType string   `json:"policy_type"`
		GroupID    *uint    `json:"group_id"`
		Value      string   `json:"value"`
		Values     []string `json:"values"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	orgID := c.GetUint("org_id")
	username, _ := c.Get("username")

	svc := &policyService.PolicyService{}

	if len(req.Values) > 0 {
		var policies []models.Policy
		for _, val := range req.Values {
			if val == "" {
				continue
			}
			policies = append(policies, models.Policy{
				Title:      req.Title,
				Category:   req.Category,
				PolicyType: req.PolicyType,
				GroupID:    req.GroupID,
				Value:      val,
			})
		}
		if err := svc.BulkCreatePoliciesArray(policies, orgID, username.(string)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lưu hàng loạt. Đã hoàn tác (rollback) toàn bộ!"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Đã gửi yêu cầu phê duyệt %d chính sách.", len(policies))})
		return
	}

	// Xử lý Single nếu không có mảng Values
	p := models.Policy{
		Title:      req.Title,
		Category:   req.Category,
		PolicyType: req.PolicyType,
		GroupID:    req.GroupID,
		Value:      req.Value,
	}

	err := svc.CreatePolicyRequest(p, orgID, username.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu phê duyệt chính sách."})
}

func BulkDeletePolicy(c *gin.Context) {
	var req struct {
		IDs    []uint `json:"ids"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	orgID := c.GetUint("org_id")
	username, _ := c.Get("username")

	svc := &policyService.PolicyService{}
	if err := svc.BulkDeletePolicyRequest(req.IDs, orgID, username.(string), req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu xóa hàng loạt chính sách."})
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
func GetAssetWhitelist(c *gin.Context) {
	hwid := c.Param("hwid")
	orgID := c.GetUint("org_id")

	svc := &policyService.PolicyService{}
	list, err := svc.GetAssetWhitelist(orgID, hwid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn Whitelist"})
		return
	}
	c.JSON(http.StatusOK, list)
}

// RemoveWhitelistItem (GATE): Admin có thể xóa một mục trong Whitelist nếu thấy nó khả nghi
func RemoveWhitelistItem(c *gin.Context) {
	idStr := c.Param("item_id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	orgID := c.GetUint("org_id")

	if err := database.DB.Where("id = ? AND org_id = ?", uint(id), orgID).Delete(&models.WhitelistItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa mục whitelist"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa mục khỏi Whitelist"})
}

type CheckViolationRequest struct {
	OrgID    uint   `json:"org_id" binding:"required"`
	HWID     string `json:"hwid" binding:"required"`
	Category string `json:"category" binding:"required"`
	Value    string `json:"value" binding:"required"`
}

// CheckPolicyViolation is an internal endpoint for other services to check for policy violations.
func CheckPolicyViolation(c *gin.Context) {
	var req CheckViolationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	svc := &policyService.PolicyService{}

	isViolation, message, err := svc.CheckPolicyViolation(req.OrgID, req.HWID, req.Category, req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi kiểm tra chính sách: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_violation": isViolation,
		"message":      message,
	})
}

type InternalBaselineReq struct {
	AssetHWID string   `json:"asset_hwid"`
	OrgID     uint     `json:"org_id"`
	Category  string   `json:"category"`
	Values    []string `json:"values"`
}

func InternalSaveBaseline(c *gin.Context) {
	var req InternalBaselineReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	svc := &policyService.PolicyService{}
	// Sử dụng hàm SaveBaselineItems đã viết ở bước trước
	err := svc.SaveBaselineItems(req.OrgID, req.AssetHWID, req.Category, req.Values)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "success"})
}
