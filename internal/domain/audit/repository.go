package audit

import "context"

// Repository 审核仓储接口
type Repository interface {
	// CreateAudit 创建审核记录
	CreateAudit(ctx context.Context, audit *AuditResult) error

	// GetAuditByID 根据ID获取审核记录
	GetAuditByID(ctx context.Context, id string) (*AuditResult, error)

	// GetAuditByReimbursementID 根据报销单ID获取审核记录
	GetAuditByReimbursementID(ctx context.Context, reimbursementID string) (*AuditResult, error)

	// UpdateAudit 更新审核记录
	UpdateAudit(ctx context.Context, audit *AuditResult) error

	// ListAudits 查询审核列表
	ListAudits(ctx context.Context, filter *AuditFilter) ([]*AuditResult, int64, error)

	// DeleteAudit 删除审核记录
	DeleteAudit(ctx context.Context, id string) error
}
