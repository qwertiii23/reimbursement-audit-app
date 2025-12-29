package audit

import (
	"context"
)

// ServiceInterface 审核服务接口
type ServiceInterface interface {
	StartAudit(ctx context.Context, reimbursementID string) (*AuditResult, error)
	GetAuditStatus(ctx context.Context, auditID string) (*AuditResult, error)
	GetAuditByReimbursementID(ctx context.Context, reimbursementID string) (*AuditResult, error)
	RetryAudit(ctx context.Context, auditID string) (*AuditResult, error)
}
