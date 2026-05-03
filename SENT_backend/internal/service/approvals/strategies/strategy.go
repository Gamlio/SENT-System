package strategies

import (
	"sent_backend/internal/models"

	"gorm.io/gorm"
)

// ApprovalStrategy là bộ khung chuẩn cho mọi loại phê duyệt
type ApprovalStrategy interface {
	OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error
	OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error
}

// Registry: Nơi đăng ký tất cả các module.
// Nếu sau này có thêm "DUYỆT NGHỈ PHÉP", chỉ cần thêm 1 dòng vào đây!
var Registry = map[string]ApprovalStrategy{
	"ASSET_ENROLL":      &assetEnrollStrategy{},
	"ASSET_DELETE":      &assetDeleteStrategy{},
	"ASSET_BULK_DELETE": &assetBulkDeleteStrategy{},

	"POLICY_CREATE":      &PolicyCreateStrategy{},
	"POLICY_DELETE":      &PolicyDeleteStrategy{},
	"POLICY_BULK_DELETE": &PolicyBulkDeleteStrategy{},

	"DOCUMENT_UPLOAD": &DocumentUploadStrategy{},
	"DOCUMENT_DELETE": &DocumentUploadStrategy{},
	"USER_CREATE":     &UserCreateStrategy{},
	"USER_UPDATE":     &UserUpdateStrategy{},
	"USER_DELETE":     &UserDeleteStrategy{},
	"GROUP_CREATE":    &GroupCreateStrategy{},
	"GROUP_UPDATE":    &GroupUpdateStrategy{},
	"GROUP_DELETE":    &GroupDeleteStrategy{},
}

// GetStrategy: Hàm lấy bộ xử lý tương ứng
func GetStrategy(moduleType string) ApprovalStrategy {
	return Registry[moduleType]
}
