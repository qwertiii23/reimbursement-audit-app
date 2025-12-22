// repository.go OCR仓储接口
// 功能点：
// 1. 定义OCR结果存储接口
// 2. 提供OCR查询方法

package ocr

import (
	"context"
)

// Repository OCR仓储接口
type Repository interface {
	// 发票相关方法
	CreateInvoice(ctx context.Context, invoice *Invoice) error
	CreateInvoices(ctx context.Context, invoices []*Invoice) error
	GetInvoiceByID(ctx context.Context, id string) (*Invoice, error)
	UpdateInvoice(ctx context.Context, invoice *Invoice) error
	DeleteInvoice(ctx context.Context, id string) error
	ListInvoicesByReimbursementID(ctx context.Context, reimbursementID string) ([]*Invoice, error)
}
