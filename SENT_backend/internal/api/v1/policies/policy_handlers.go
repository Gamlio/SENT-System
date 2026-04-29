package policies

import (
	"fmt"
	"net/http"
	"sent_backend/internal/database"
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

// AddPolicy (GATE): Tiếp nhận yêu cầu thêm luật lẻ
func AddPolicy(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Đã gửi yêu cầu xóa %d chính sách.", len(req.IDs))})
}

// SyncPoliciesForAsset (GATE): Cổng đồng bộ cho asset
func SyncPoliciesForAsset(c *gin.Context) {
	orgID := c.GetUint("org_id")

	var req struct {
		AssetHWID string `json:"asset_hwid"`
		Version   int64  `json:"version"` // Agent gửi version hiện tại lên
	}

	// Đọc từ JSON payload để khớp với Agent
	if err := c.ShouldBindJSON(&req); err != nil || req.AssetHWID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu thông tin AssetHWID hoặc payload không hợp lệ"})
		return
	}

	svc := &policyService.PolicyService{}
	policies, err := svc.GetEffectivePolicies(orgID, req.AssetHWID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Tính toán LastUpdated lớn nhất để làm Version
	var maxUpdated int64 = 0
	for _, p := range policies {
		if p.UpdatedAt.Unix() > maxUpdated {
			maxUpdated = p.UpdatedAt.Unix()
		}
	}

	// Nếu version của Agent >= version mới nhất ở Backend -> Không cần tải lại (Tiết kiệm băng thông)
	if req.Version > 0 && req.Version >= maxUpdated {
		c.JSON(http.StatusOK, gin.H{"status": "not_modified"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "synchronized",
		"version":  maxUpdated,
		"policies": policies,
	})
}

// GetPolicyGroups: Lấy danh sách các nhóm chính sách đã tạo
func GetPolicyGroups(c *gin.Context) {
	orgID := c.GetUint("org_id")
	var groups []models.PolicyGroup

	// Lấy tất cả nhóm thuộc tổ chức này
	if err := database.DB.Where("org_id = ?", orgID).Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách nhóm"})
		return
	}

	c.JSON(http.StatusOK, groups)
}
