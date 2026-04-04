package policies

import (
	"fmt"
	"net/http"
	"sent_backend/internal/models"
	policyService "sent_backend/internal/service/policies"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetPoliciesByCategory (GATE): Truy vấn danh sách
func GetPoliciesByCategory(c *gin.Context) {
	orgID := c.GetUint("org_id")
	category := c.Query("category")
	status := c.Query("status")

	svc := &policyService.PolicyService{}
	list := svc.GetPolicies(orgID, category, status)
	c.JSON(http.StatusOK, list)
}

// AddUniversalPolicy (GATE): Tiếp nhận yêu cầu thêm luật lẻ
func AddUniversalPolicy(c *gin.Context) {
	var req models.UniversalPolicy
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

// AddBulkPolicies (GATE): Tiếp nhận nạp luật hàng loạt
func AddBulkPolicies(c *gin.Context) {
	var req struct {
		Policies []policyService.ExcelRow `json:"policies"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu form không hợp lệ"})
		return
	}

	orgID := c.GetUint("org_id")
	svc := &policyService.PolicyService{}
	count, err := svc.BulkCreatePolicyRequest(req.Policies, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã nạp luật thành công, vui lòng chờ duyệt", "count": count})
}

// DeletePolicy (GATE): Gửi yêu cầu xóa 1 luật
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

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu xóa chính sách, vui lòng chờ duyệt."})
}

// DeleteBulkPolicies (GATE): Gửi yêu cầu xóa nhiều luật
func DeleteBulkPolicies(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	orgID := c.GetUint("org_id")
	username, _ := c.Get("username")

	svc := &policyService.PolicyService{}
	if err := svc.BulkDeletePolicyRequest(req.IDs, orgID, username.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Đã gửi yêu cầu xóa %d chính sách.", len(req.IDs))})
}

// SyncPoliciesForAsset (GATE): Cổng đồng bộ cho asset
func SyncPoliciesForAsset(c *gin.Context) {
	orgID := c.GetUint("org_id")
	hwid := c.Query("hwid")
	if hwid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu thông tin HWID"})
		return
	}

	svc := &policyService.PolicyService{}
	policies, err := svc.GetEffectivePolicies(orgID, hwid)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "synchronized", "policies": policies})
}
