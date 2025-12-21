// repository.go 报销单仓储接口定义
// 功能点：
// 1. 定义报销单仓储接口
// 2. 定义发票仓储接口
// 3. 定义审核结果仓储接口
// 4. 提供数据访问抽象层
// 5. 支持事务管理
// 6. 支持查询和分页

package reimbursement

import (
	"context"
)

// Repository 报销单仓储接口
type Repository interface {
	// 报销单相关方法
	CreateReimbursement(ctx context.Context, reimbursement *Reimbursement) error
	GetReimbursementByID(ctx context.Context, id string) (*Reimbursement, error)
	UpdateReimbursement(ctx context.Context, reimbursement *Reimbursement) error
	DeleteReimbursement(ctx context.Context, id string) error
	ListReimbursementsByUserID(ctx context.Context, userID string, page, size int) ([]*Reimbursement, int64, error)
	ListReimbursementsByDateRange(ctx context.Context, startDate, endDate string, page, size int) ([]*Reimbursement, int64, error)
	ListReimbursementsByStatus(ctx context.Context, status string, page, size int) ([]*Reimbursement, int64, error)
	SearchReimbursements(ctx context.Context, keyword string, page, size int) ([]*Reimbursement, int64, error)

	// 发票相关方法
	CreateInvoice(ctx context.Context, invoice *Invoice) error
	CreateInvoices(ctx context.Context, invoices []*Invoice) error
	GetInvoiceByID(ctx context.Context, id string) (*Invoice, error)
	UpdateInvoice(ctx context.Context, invoice *Invoice) error
	DeleteInvoice(ctx context.Context, id string) error
	ListInvoicesByReimbursementID(ctx context.Context, reimbursementID string) ([]*Invoice, error)

	// 审核结果相关方法
	CreateAuditResult(ctx context.Context, result *AuditResult) error
	GetAuditResultByID(ctx context.Context, id string) (*AuditResult, error)
	GetLatestAuditResultByReimbursementID(ctx context.Context, reimbursementID string) (*AuditResult, error)
	UpdateAuditResult(ctx context.Context, result *AuditResult) error
	ListAuditResultsByReimbursementID(ctx context.Context, reimbursementID string) ([]*AuditResult, error)

	// 审核状态相关方法
	CreateAuditStatus(ctx context.Context, status *AuditStatus) error
	GetAuditStatusByReimbursementID(ctx context.Context, reimbursementID string) (*AuditStatus, error)
	UpdateAuditStatus(ctx context.Context, status *AuditStatus) error

	// 事务相关方法
	BeginTx(ctx context.Context) (Tx, error)
}

// Tx 事务接口
type Tx interface {
	// 提交事务
	Commit() error
	// 回滚事务
	Rollback() error
	// 获取事务上下文
	Context() context.Context
}
