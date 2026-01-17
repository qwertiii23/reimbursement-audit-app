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
	ManualAudit(ctx context.Context, auditID string, action string, reason string, operatorID string, operatorName string, ipAddress string) (*AuditResult, error)
	GetFlowLogsByReimbursementID(ctx context.Context, reimbursementID string) ([]*AuditFlowLog, error)
	GetFlowLogsByAuditID(ctx context.Context, auditID string) ([]*AuditFlowLog, error)
}
