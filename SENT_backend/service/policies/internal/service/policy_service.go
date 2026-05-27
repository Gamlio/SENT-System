package policies

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"encoding/json"
	"errors"
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
		return nil, fmt.Errorf("thiết bị không tồn tại: %w", err)
	}

	var effectivePolicies []models.Policy

	// Xây dựng câu truy vấn gom cụm điều kiện tối ưu để tránh chia nhỏ câu lệnh SQL xuống database
	query := database.DB.Where("org_id = ? AND is_active = ? AND approval_status = ?", orgID, true, "APPROVED")

	if asset.GroupID != nil {
		// Tìm luật thỏa mãn: (Không gán nhóm + Không gán máy) HOẶC (Thuộc nhóm máy này) HOẶC (Thuộc riêng máy này)
		query = query.Where("(group_id IS NULL AND asset_hwid = '') OR (group_id = ?) OR (asset_hwid = ?)", *asset.GroupID, hwid)
	} else {
		// Tìm luật thỏa mãn: (Không gán nhóm + Không gán máy) HOẶC (Thuộc riêng máy này)
		query = query.Where("(group_id IS NULL AND asset_hwid = '') OR (asset_hwid = ?)", hwid)
	}

	if err := query.Order("created_at DESC").Find(&effectivePolicies).Error; err != nil {
		return nil, fmt.Errorf("lỗi truy vấn luật hiệu dụng: %w", err)
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
// [TỐI ƯU] Hàm được viết lại để truy vấn trực tiếp vào DB thay vì tải toàn bộ luật về xử lý.
func (s *PolicyService) CheckPolicyViolation(orgID uint, hwid string, category string, value string) (bool, string, error) {
	// Lấy thông tin group của asset để xác định phạm vi luật
	var asset models.Asset
	if err := database.DB.Where("asset_hwid = ? AND org_id = ?", hwid, orgID).Select("group_id").First(&asset).Error; err != nil {
		// Nếu không tìm thấy asset, không thể kiểm tra, coi như không vi phạm nhưng log lỗi
		return false, "", fmt.Errorf("không thể kiểm tra vi phạm, không tìm thấy asset %s: %w", hwid, err)
	}

	cleanValue := strings.ToLower(strings.TrimSpace(value))
	cleanCategory := strings.ToUpper(category)

	var matchingPolicy models.Policy

	// === BƯỚC 1: Tối ưu truy vấn tìm kiếm một luật khớp chính xác với giá trị ===
	query := database.DB.Where(
		"org_id = ? AND is_active = ? AND approval_status = ? AND category = ? AND LOWER(value) = ?",
		orgID, true, "APPROVED", cleanCategory, cleanValue,
	)

	// Áp dụng phạm vi (Scope) của luật: (Toàn cục) HOẶC (Nhóm) HOẶC (Cá nhân)
	if asset.GroupID != nil {
		query = query.Where("(group_id IS NULL AND asset_hwid = '') OR (group_id = ?) OR (asset_hwid = ?)", *asset.GroupID, hwid)
	} else {
		query = query.Where("(group_id IS NULL AND asset_hwid = '') OR (asset_hwid = ?)", hwid)
	}

	err := query.Select("id, title, policy_type").First(&matchingPolicy).Error

	// Trường hợp 1: Tìm thấy một luật khớp chính xác (Whitelist hoặc Blacklist)
	if err == nil {
		if matchingPolicy.PolicyType == "BLACKLIST" {
			return true, fmt.Sprintf("Phát hiện vi phạm Blacklist: %s", matchingPolicy.Title), nil
		}
		// Nếu khớp với một luật Whitelist -> Hợp lệ, không vi phạm.
		return false, "", nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, "", fmt.Errorf("lỗi DB khi kiểm tra luật khớp chính xác: %w", err)
	}

	// === BƯỚC 2: Kiểm tra logic Zero Trust (nếu không có luật nào khớp chính xác) ===
	var hasWhitelistPolicy models.Policy
	whitelistCheckQuery := database.DB.Where(
		"org_id = ? AND is_active = ? AND approval_status = ? AND category = ? AND policy_type = 'WHITELIST'",
		orgID, true, "APPROVED", cleanCategory,
	)

	if asset.GroupID != nil {
		whitelistCheckQuery = whitelistCheckQuery.Where("(group_id IS NULL AND asset_hwid = '') OR (group_id = ?) OR (asset_hwid = ?)", *asset.GroupID, hwid)
	} else {
		whitelistCheckQuery = whitelistCheckQuery.Where("(group_id IS NULL AND asset_hwid = '') OR (asset_hwid = ?)", hwid)
	}

	if err = whitelistCheckQuery.Select("id").First(&hasWhitelistPolicy).Error; err == nil {
		return true, fmt.Sprintf("Vi phạm Whitelist: [%s] không được phép hoạt động (Zero Trust)", value), nil
	}

	return false, "", nil
}
func (s *PolicyService) SaveBaselineItems(orgID uint, hwid string, category string, values []string) error {
	if len(values) == 0 {
		return nil
	}

	// 1. Kiểm tra nhóm (GroupID) hiện tại của máy trạm để gán thừa kế (nếu có)
	var asset models.Asset
	if err := database.DB.Where("asset_hwid = ? AND org_id = ?", hwid, orgID).First(&asset).Error; err != nil {
		return fmt.Errorf("không thể tìm thấy thông tin thiết bị để nạp baseline: %w", err)
	}

	var items []models.Policy
	for _, val := range values {
		cleanVal := strings.TrimSpace(val)
		if cleanVal == "" {
			continue
		}

		// Chuẩn hóa tiêu đề hiển thị trên UI cho SOC dễ quản lý
		title := fmt.Sprintf("Baseline [%s]: %s", hwid, cleanVal)
		if len(title) > 150 {
			title = title[:147] + "..."
		}

		items = append(items, models.Policy{
			OrgID:          orgID,
			Title:          title,
			Category:       strings.ToUpper(category), // Đồng bộ HOA để Agent map chuẩn O(1)
			Value:          strings.ToLower(cleanVal), // Đồng bộ thường cho mã băm/tiến trình
			PolicyType:     "WHITELIST",               // Baseline bản chất là danh sách trắng
			ApprovalStatus: "APPROVED",                // Tự động phê duyệt vì là máy sạch ban đầu
			IsActive:       true,
			CreatedBy:      "System_Baseline_Engine",
			ApprovedBy:     "System_Auto_Sign",
			GroupID:        asset.GroupID, // Thừa kế nhóm để tối ưu phân vùng
			AssetHWID:      hwid,          // Khóa định danh máy sở hữu luật
		})
	}

	// 2. Sử dụng mệnh đề OnConflict của Postgres chống tràn bộ nhớ dữ liệu lặp
	return database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "org_id"}, {Name: "category"}, {Name: "policy_type"}, {Name: "value"}, {Name: "asset_hwid"}},
		DoNothing: true, // Nếu trùng bản ghi máy sạch cũ thì bỏ qua không ghi đè
	}).CreateInBatches(items, 100).Error
}
