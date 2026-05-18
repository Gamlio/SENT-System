package policies

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PolicyService struct{}

type ExcelRow struct {
	Title      string `json:"title"`
	Category   string `json:"category"`
	Value      string `json:"value"`
	PolicyType string `json:"policy_type"`
}

// CreatePolicyRequest: Logic tạo luật và Ticket phê duyệt
func (s *PolicyService) CreatePolicyRequest(req models.Policy, orgID uint, creator string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		req.OrgID = orgID
		req.ApprovalStatus = "PENDING"
		req.CreatedBy = creator
		req.IsActive = false // Chưa được duyệt thì chưa kích hoạt

		if err := tx.Create(&req).Error; err != nil {
			return err
		}

		// Tạo vé phê duyệt tập trung
		ticket := models.ApprovalTicket{
			OrgID:        orgID,
			ModuleType:   "POLICY_CREATE",
			ActionType:   "CREATE",
			TargetID:     req.ID,
			TargetName:   fmt.Sprintf("[%s] %s", req.Category, req.Value),
			Status:       "PENDING",
			RequestedBy:  creator,
			SnapshotData: fmt.Sprintf(`{"title": "%s", "value": "%s"}`, req.Title, req.Value),
		}
		return tx.Create(&ticket).Error
	})
}

// BulkCreatePoliciesArray: Tạo hàng loạt luật từ mảng với Transaction (Hỗ trợ Rollback)
func (s *PolicyService) BulkCreatePoliciesArray(policies []models.Policy, orgID uint, creator string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		for _, req := range policies {
			req.OrgID = orgID
			req.ApprovalStatus = "PENDING"
			req.CreatedBy = creator
			req.IsActive = false // Chưa được duyệt thì chưa kích hoạt

			if err := tx.Create(&req).Error; err != nil {
				return err // Gặp lỗi ở 1 dòng sẽ tự động Rollback toàn bộ
			}

			// Tạo vé phê duyệt
			ticket := models.ApprovalTicket{
				OrgID:        orgID,
				ModuleType:   "POLICY_CREATE",
				ActionType:   "CREATE",
				TargetID:     req.ID,
				TargetName:   fmt.Sprintf("[%s] %s", req.Category, req.Value),
				Status:       "PENDING",
				RequestedBy:  creator,
				SnapshotData: fmt.Sprintf(`{"title": "%s", "value": "%s"}`, req.Title, req.Value),
			}
			if err := tx.Create(&ticket).Error; err != nil {
				return err // Lỗi tạo vé -> cũng Rollback toàn bộ
			}
		}
		return nil
	})
}

// BulkCreatePolicyRequest: Xử lý nạp luật hàng loạt từ Excel
func (s *PolicyService) BulkCreatePolicyRequest(rows []ExcelRow, orgID uint) (int, error) {
	count := 0
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, row := range rows {
			if strings.TrimSpace(row.Value) == "" {
				continue
			}

			policy := models.Policy{
				OrgID:          orgID,
				Title:          row.Title,
				Category:       row.Category,
				Value:          row.Value,
				PolicyType:     row.PolicyType,
				ApprovalStatus: "PENDING",
				CreatedBy:      "System_Import",
			}

			if err := tx.Create(&policy).Error; err != nil {
				return err
			}

			ticket := models.ApprovalTicket{
				OrgID:       orgID,
				ModuleType:  "POLICY_CREATE",
				TargetID:    policy.ID,
				TargetName:  fmt.Sprintf("[%s] %s", policy.Category, policy.Value),
				Status:      "PENDING",
				RequestedBy: "System_Import",
			}
			if err := tx.Create(&ticket).Error; err != nil {
				return err
			}
			count++
		}
		return nil
	})
	return count, err
}

// GetEffectivePolicies: Engine tính toán luật hiệu dụng cho asset
func (s *PolicyService) GetEffectivePolicies(orgID uint, hwid string) ([]models.Policy, error) {
	var asset models.Asset
	if err := database.DB.Where("asset_hwid = ? AND org_id = ?", hwid, orgID).First(&asset).Error; err != nil {
		return nil, err
	}

	var effectivePolicies []models.Policy

	// 1. Lấy Global Policies (GroupID IS NULL)
	var globalPolicies []models.Policy
	database.DB.Where("org_id = ? AND group_id IS NULL AND is_active = true AND approval_status = ?", orgID, "APPROVED").Find(&globalPolicies)
	effectivePolicies = append(effectivePolicies, globalPolicies...)

	// 2. Lấy Group Policies nếu máy trạm có thuộc nhóm nào đó
	if asset.GroupID != nil {
		var groupPolicies []models.Policy
		database.DB.Where("org_id = ? AND group_id = ? AND is_active = true AND approval_status = ?", orgID, *asset.GroupID, "APPROVED").Find(&groupPolicies)
		effectivePolicies = append(effectivePolicies, groupPolicies...)
	}

	return effectivePolicies, nil
}

// GetPolicies: Truy vấn danh sách có lọc
func (s *PolicyService) GetPolicies(orgID uint, category, status string) []models.Policy {
	var list []models.Policy
	query := database.DB.Where("org_id = ?", orgID)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if status != "" {
		query = query.Where("approval_status = ?", status)
	}
	query.Order("created_at desc").Find(&list)
	return list
}

// DeletePolicyRequest: Tạo yêu cầu xóa 1 luật
func (s *PolicyService) DeletePolicyRequest(id uint, orgID uint, creator string) error {
	var policy models.Policy
	if err := database.DB.Where("id = ? AND org_id = ?", id, orgID).First(&policy).Error; err != nil {
		return fmt.Errorf("không tìm thấy chính sách")
	}

	// Tạo vé phê duyệt loại DELETE
	ticket := models.ApprovalTicket{
		OrgID:        orgID,
		ModuleType:   "POLICY_DELETE",
		ActionType:   "DELETE",
		TargetID:     id,
		TargetName:   fmt.Sprintf("Xóa luật: %s", policy.Title),
		Status:       "PENDING",
		RequestedBy:  creator,
		SnapshotData: fmt.Sprintf(`{"id": %d, "title": "%s"}`, id, policy.Title),
	}
	return database.DB.Create(&ticket).Error
}

// BulkDeletePolicyRequest: Tạo yêu cầu xóa hàng loạt luật
func (s *PolicyService) BulkDeletePolicyRequest(ids []uint, orgID uint, creator string, reason string) error {
	if len(ids) == 0 {
		return fmt.Errorf("danh sách ID trống")
	}

	// Nếu user không nhập reason, đặt một lý do mặc định
	if reason == "" {
		reason = "Yêu cầu xóa hàng loạt từ Admin"
	}

	// Gói danh sách ID vào Snapshot để Strategy sau này xử lý
	snapshotData, _ := json.Marshal(map[string]interface{}{
		"ids":    ids,
		"reason": reason,
	})

	ticket := models.ApprovalTicket{
		OrgID:         orgID,
		ModuleType:    "POLICY_BULK_DELETE",
		ActionType:    "DELETE",
		TargetID:      0,
		TargetName:    fmt.Sprintf("Xóa %d chính sách", len(ids)),
		Status:        "PENDING",
		RequestedBy:   creator,
		RequestReason: reason,
		SnapshotData:  string(snapshotData),
	}

	return database.DB.Create(&ticket).Error
}

// CheckPolicyViolation: Bộ lọc thông minh so khớp dữ liệu log với Policy (Whitelist/Blacklist)
func (s *PolicyService) CheckPolicyViolation(orgID uint, hwid string, category string, value string) (bool, string, error) {
	policies, err := s.GetEffectivePolicies(orgID, hwid)
	if err != nil {
		return false, "", err
	}

	hasWhitelist := false
	inWhitelist := false

	for _, p := range policies {
		// Chỉ kiểm tra các policy cùng danh mục (VD: SOFTWARE, USB, NETWORK)
		if p.Category != category {
			continue
		}

		// So khớp giá trị (không phân biệt chữ hoa chữ thường)
		match := strings.EqualFold(strings.TrimSpace(p.Value), strings.TrimSpace(value))

		if p.PolicyType == "BLACKLIST" && match {
			return true, fmt.Sprintf("Phát hiện vi phạm Blacklist: %s", p.Title), nil
		}

		if p.PolicyType == "WHITELIST" {
			hasWhitelist = true
			if match {
				inWhitelist = true
			}
		}
	}

	// Nếu có áp dụng Whitelist cho danh mục này mà giá trị log không khớp -> Vi phạm
	if hasWhitelist && !inWhitelist {
		return true, fmt.Sprintf("Vi phạm Whitelist: [%s] không được phép hoạt động", value), nil
	}

	return false, "", nil
}
func (s *PolicyService) SaveBaselineItems(orgID uint, hwid string, category string, values []string) error {
	if len(values) == 0 {
		return nil
	}

	var items []models.WhitelistItem
	for _, val := range values {
		if strings.TrimSpace(val) == "" {
			continue
		}
		items = append(items, models.WhitelistItem{
			OrgID:       orgID,
			AssetHWID:   hwid,
			Type:        category, // "SOFTWARE_HASH" hoặc "PORT"
			Value:       val,
			Description: "Tự động tạo từ Baseline (1h đầu)",
		})
	}

	// [TỐI ƯU] Sử dụng OnConflict để tránh trùng lặp nếu máy trạm gửi log nhiều lần trong 1h
	return database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "org_id"}, {Name: "asset_hwid"}, {Name: "type"}, {Name: "value"}},
		DoNothing: true,
	}).CreateInBatches(items, 100).Error
}

// GetAssetWhitelist: Lấy danh sách whitelist riêng biệt của một máy trạm
func (s *PolicyService) GetAssetWhitelist(orgID uint, hwid string) ([]models.WhitelistItem, error) {
	var list []models.WhitelistItem
	err := database.DB.Where("org_id = ? AND asset_hwid = ?", orgID, hwid).Find(&list).Error
	return list, err
}
