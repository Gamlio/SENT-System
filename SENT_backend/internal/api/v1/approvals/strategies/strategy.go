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
	"AGENT_ENROLL":      &AgentEnrollStrategy{},
	"AGENT_DELETE":      &AgentDeleteStrategy{},
	"AGENT_BULK_DELETE": &AgentBulkDeleteStrategy{},
	"POLICY_CREATE":     &PolicyCreateStrategy{},
	"DOCUMENT_UPLOAD":   &DocumentUploadStrategy{},
	"USER_CREATE":       &UserCreateStrategy{},
	"USER_UPDATE":       &UserUpdateStrategy{},
	"USER_DELETE":       &UserDeleteStrategy{},
}

// GetStrategy: Hàm lấy bộ xử lý tương ứng
func GetStrategy(moduleType string) ApprovalStrategy {
	return Registry[moduleType]
}
