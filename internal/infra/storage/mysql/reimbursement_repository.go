// mysql_repository.go MySQL仓储实现
// 功能点：
// 1. 实现报销单仓储接口
// 2. 实现发票仓储接口
// 3. 实现审核结果仓储接口
// 4. 提供MySQL数据访问实现
// 5. 支持事务管理
// 6. 支持查询和分页

package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"reimbursement-audit/internal/api/middleware"
	"reimbursement-audit/internal/domain/reimbursement"

	"gorm.io/gorm"
)

// ReimbursementRepository 报销单MySQL仓储实现
type ReimbursementRepository struct {
	client *Client
}

// NewReimbursementRepository 创建报销单MySQL仓储实例
func NewReimbursementRepository(client *Client) reimbursement.Repository {
	return &ReimbursementRepository{
		client: client,
	}
}

// CreateReimbursement 创建报销单
func (r *ReimbursementRepository) CreateReimbursement(ctx context.Context, reimbursement *reimbursement.Reimbursement) error {
	// 获取traceId用于日志追踪
	traceId := middleware.GetTraceIdFromContext(ctx)

	// 使用GORM创建报销单记录
	result := r.client.GetDB().WithContext(ctx).Create(reimbursement)
	if result.Error != nil {
		return fmt.Errorf("创建报销单失败: %w, traceId: %s", result.Error, traceId)
	}

	return nil
}

// GetReimbursementByID 根据ID获取报销单
func (r *ReimbursementRepository) GetReimbursementByID(ctx context.Context, id string) (*reimbursement.Reimbursement, error) {
	var reimbursement reimbursement.Reimbursement

	// 使用GORM查询报销单
	result := r.client.GetDB().WithContext(ctx).Where("id = ?", id).First(&reimbursement)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("报销单不存在")
		}
		return nil, fmt.Errorf("获取报销单失败: %w", result.Error)
	}

	// 获取关联的发票列表
	invoices, err := r.ListInvoicesByReimbursementID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取报销单发票列表失败: %w", err)
	}
	reimbursement.Invoices = invoices

	return &reimbursement, nil
}

// UpdateReimbursement 更新报销单
func (r *ReimbursementRepository) UpdateReimbursement(ctx context.Context, reimbursement *reimbursement.Reimbursement) error {
	// 使用GORM更新报销单
	result := r.client.GetDB().WithContext(ctx).Model(reimbursement).
		Where("id = ?", reimbursement.ID).
		Updates(map[string]interface{}{
			"user_id":      reimbursement.UserID,
			"user_name":    reimbursement.UserName,
			"department":   reimbursement.Department,
			"type":         reimbursement.Type,
			"title":        reimbursement.Title,
			"description":  reimbursement.Description,
			"total_amount": reimbursement.TotalAmount,
			"currency":     reimbursement.Currency,
			"apply_date":   reimbursement.ApplyDate,
			"expense_date": reimbursement.ExpenseDate,
			"status":       reimbursement.Status,
			"updated_at":   time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("更新报销单失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("报销单不存在")
	}

	return nil
}

// DeleteReimbursement 删除报销单
func (r *ReimbursementRepository) DeleteReimbursement(ctx context.Context, id string) error {
	// 使用GORM删除报销单
	result := r.client.GetDB().WithContext(ctx).Where("id = ?", id).Delete(&reimbursement.Reimbursement{})

	if result.Error != nil {
		return fmt.Errorf("删除报销单失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("报销单不存在")
	}

	return nil
}

// GetReimbursementsByUserID 根据用户ID获取报销单列表
func (r *ReimbursementRepository) GetReimbursementsByUserID(ctx context.Context, userID string, limit, offset int) ([]*reimbursement.Reimbursement, error) {
	var reimbursements []*reimbursement.Reimbursement

	// 使用GORM查询报销单列表
	result := r.client.GetDB().WithContext(ctx).
		Where("user_id = ?", userID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&reimbursements)

	if result.Error != nil {
		return nil, fmt.Errorf("查询用户报销单失败: %w", result.Error)
	}

	return reimbursements, nil
}

// ListReimbursementsByUserID 根据用户ID获取报销单列表
func (r *ReimbursementRepository) ListReimbursementsByUserID(ctx context.Context, userID string, page, size int) ([]*reimbursement.Reimbursement, int64, error) {
	// 获取总数
	var total int64
	countResult := r.client.GetDB().WithContext(ctx).Model(&reimbursement.Reimbursement{}).Where("user_id = ?", userID).Count(&total)
	if countResult.Error != nil {
		return nil, 0, fmt.Errorf("获取报销单总数失败: %w", countResult.Error)
	}

	// 获取分页数据
	offset := (page - 1) * size
	var reimbursements []*reimbursement.Reimbursement
	result := r.client.GetDB().WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&reimbursements)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("获取报销单列表失败: %w", result.Error)
	}

	// 获取关联的发票列表
	for _, reimbursement := range reimbursements {
		invoices, err := r.ListInvoicesByReimbursementID(ctx, reimbursement.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("获取报销单发票列表失败: %w", err)
		}
		reimbursement.Invoices = invoices
	}

	return reimbursements, total, nil
}

// ListReimbursementsByDateRange 根据日期范围获取报销单列表
func (r *ReimbursementRepository) ListReimbursementsByDateRange(ctx context.Context, startDate, endDate string, page, size int) ([]*reimbursement.Reimbursement, int64, error) {
	// 获取总数
	var total int64
	countResult := r.client.GetDB().WithContext(ctx).Model(&reimbursement.Reimbursement{}).
		Where("apply_date BETWEEN ? AND ?", startDate, endDate).
		Count(&total)

	if countResult.Error != nil {
		return nil, 0, fmt.Errorf("获取报销单总数失败: %w", countResult.Error)
	}

	// 获取分页数据
	offset := (page - 1) * size
	var reimbursements []*reimbursement.Reimbursement
	result := r.client.GetDB().WithContext(ctx).
		Where("apply_date BETWEEN ? AND ?", startDate, endDate).
		Order("apply_date DESC").
		Limit(size).
		Offset(offset).
		Find(&reimbursements)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("获取报销单列表失败: %w", result.Error)
	}

	// 获取关联的发票列表
	for _, reimbursement := range reimbursements {
		invoices, err := r.ListInvoicesByReimbursementID(ctx, reimbursement.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("获取报销单发票列表失败: %w", err)
		}
		reimbursement.Invoices = invoices
	}

	return reimbursements, total, nil
}

// ListReimbursementsByStatus 根据状态获取报销单列表
func (r *ReimbursementRepository) ListReimbursementsByStatus(ctx context.Context, status string, page, size int) ([]*reimbursement.Reimbursement, int64, error) {
	// 获取总数
	var total int64
	countResult := r.client.GetDB().WithContext(ctx).Model(&reimbursement.Reimbursement{}).
		Where("status = ?", status).
		Count(&total)

	if countResult.Error != nil {
		return nil, 0, fmt.Errorf("获取报销单总数失败: %w", countResult.Error)
	}

	// 获取分页数据
	offset := (page - 1) * size
	var reimbursements []*reimbursement.Reimbursement
	result := r.client.GetDB().WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&reimbursements)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("获取报销单列表失败: %w", result.Error)
	}

	// 获取关联的发票列表
	for _, reimbursement := range reimbursements {
		invoices, err := r.ListInvoicesByReimbursementID(ctx, reimbursement.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("获取报销单发票列表失败: %w", err)
		}
		reimbursement.Invoices = invoices
	}

	return reimbursements, total, nil
}

// SearchReimbursements 搜索报销单
func (r *ReimbursementRepository) SearchReimbursements(ctx context.Context, keyword string, page, size int) ([]*reimbursement.Reimbursement, int64, error) {
	// 获取总数
	var total int64
	searchPattern := "%" + keyword + "%"
	countResult := r.client.GetDB().WithContext(ctx).Model(&reimbursement.Reimbursement{}).
		Where("user_name LIKE ? OR title LIKE ? OR description LIKE ?", searchPattern, searchPattern, searchPattern).
		Count(&total)

	if countResult.Error != nil {
		return nil, 0, fmt.Errorf("获取报销单总数失败: %w", countResult.Error)
	}

	// 获取分页数据
	offset := (page - 1) * size
	var reimbursements []*reimbursement.Reimbursement
	result := r.client.GetDB().WithContext(ctx).
		Where("user_name LIKE ? OR title LIKE ? OR description LIKE ?", searchPattern, searchPattern, searchPattern).
		Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&reimbursements)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("获取报销单列表失败: %w", result.Error)
	}

	// 获取关联的发票列表
	for _, reimbursement := range reimbursements {
		invoices, err := r.ListInvoicesByReimbursementID(ctx, reimbursement.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("获取报销单发票列表失败: %w", err)
		}
		reimbursement.Invoices = invoices
	}

	return reimbursements, total, nil
}

// CreateInvoice 创建发票
func (r *ReimbursementRepository) CreateInvoice(ctx context.Context, invoice *reimbursement.Invoice) error {
	// 获取traceId用于日志追踪
	traceId := middleware.GetTraceIdFromContext(ctx)

	// 使用GORM创建发票记录
	result := r.client.GetDB().WithContext(ctx).Create(invoice)
	if result.Error != nil {
		return fmt.Errorf("创建发票失败: %w, traceId: %s", result.Error, traceId)
	}

	return nil
}

// CreateInvoices 批量创建发票
func (r *ReimbursementRepository) CreateInvoices(ctx context.Context, invoices []*reimbursement.Invoice) error {
	if len(invoices) == 0 {
		return nil
	}

	// 获取traceId用于日志追踪
	traceId := middleware.GetTraceIdFromContext(ctx)

	// 使用GORM批量创建发票记录
	result := r.client.GetDB().WithContext(ctx).CreateInBatches(invoices, 100) // 每批最多100条
	if result.Error != nil {
		return fmt.Errorf("批量创建发票失败: %w, traceId: %s", result.Error, traceId)
	}

	return nil
}

// GetInvoiceByID 根据ID获取发票
func (r *ReimbursementRepository) GetInvoiceByID(ctx context.Context, id string) (*reimbursement.Invoice, error) {
	var invoice reimbursement.Invoice

	// 使用GORM查询发票
	result := r.client.GetDB().WithContext(ctx).Where("id = ?", id).First(&invoice)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("发票不存在")
		}
		return nil, fmt.Errorf("查询发票失败: %w", result.Error)
	}

	return &invoice, nil
}

// UpdateInvoice 更新发票
func (r *ReimbursementRepository) UpdateInvoice(ctx context.Context, invoice *reimbursement.Invoice) error {
	// 使用GORM更新发票
	result := r.client.GetDB().WithContext(ctx).Model(invoice).
		Where("id = ?", invoice.ID).
		Updates(map[string]interface{}{
			"reimbursement_id": invoice.ReimbursementID,
			"type":             invoice.Type,
			"code":             invoice.Code,
			"number":           invoice.Number,
			"date":             invoice.Date,
			"amount":           invoice.Amount,
			"tax_amount":       invoice.TaxAmount,
			"payer":            invoice.Payer,
			"payee":            invoice.Payee,
			"buyer_name":       invoice.BuyerName,
			"buyer_tax_no":     invoice.BuyerTaxNo,
			"seller_name":      invoice.SellerName,
			"seller_tax_no":    invoice.SellerTaxNo,
			"commodity_name":   invoice.CommodityName,
			"specification":    invoice.Specification,
			"unit":             invoice.Unit,
			"quantity":         invoice.Quantity,
			"price":            invoice.Price,
			"image_path":       invoice.ImagePath,
			"ocr_result":       invoice.OCRResult,
			"status":           invoice.Status,
			"updated_at":       invoice.UpdatedAt,
		})

	if result.Error != nil {
		return fmt.Errorf("更新发票失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("发票不存在")
	}

	return nil
}

// DeleteInvoice 删除发票
func (r *ReimbursementRepository) DeleteInvoice(ctx context.Context, id string) error {
	// 使用GORM删除发票
	result := r.client.GetDB().WithContext(ctx).Where("id = ?", id).Delete(&reimbursement.Invoice{})

	if result.Error != nil {
		return fmt.Errorf("删除发票失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("发票不存在")
	}

	return nil
}

// ListInvoicesByReimbursementID 根据报销单ID获取发票列表
func (r *ReimbursementRepository) ListInvoicesByReimbursementID(ctx context.Context, reimbursementID string) ([]*reimbursement.Invoice, error) {
	var invoices []*reimbursement.Invoice

	// 使用GORM查询发票列表
	result := r.client.GetDB().WithContext(ctx).
		Where("reimbursement_id = ?", reimbursementID).
		Order("created_at ASC").
		Find(&invoices)

	if result.Error != nil {
		return nil, fmt.Errorf("获取发票列表失败: %w", result.Error)
	}

	return invoices, nil
}

// CreateAuditResult 创建审核结果
func (r *ReimbursementRepository) CreateAuditResult(ctx context.Context, result *reimbursement.AuditResult) error {
	// TODO: 实现创建审核结果逻辑
	return nil
}

// GetAuditResultByID 根据ID获取审核结果
func (r *ReimbursementRepository) GetAuditResultByID(ctx context.Context, id string) (*reimbursement.AuditResult, error) {
	// TODO: 实现获取审核结果逻辑
	return nil, nil
}

// GetLatestAuditResultByReimbursementID 根据报销单ID获取最新审核结果
func (r *ReimbursementRepository) GetLatestAuditResultByReimbursementID(ctx context.Context, reimbursementID string) (*reimbursement.AuditResult, error) {
	// TODO: 实现获取最新审核结果逻辑
	return nil, nil
}

// UpdateAuditResult 更新审核结果
func (r *ReimbursementRepository) UpdateAuditResult(ctx context.Context, result *reimbursement.AuditResult) error {
	// TODO: 实现更新审核结果逻辑
	return nil
}

// ListAuditResultsByReimbursementID 根据报销单ID获取审核结果列表
func (r *ReimbursementRepository) ListAuditResultsByReimbursementID(ctx context.Context, reimbursementID string) ([]*reimbursement.AuditResult, error) {
	// TODO: 实现获取审核结果列表逻辑
	return nil, nil
}

// CreateAuditStatus 创建审核状态
func (r *ReimbursementRepository) CreateAuditStatus(ctx context.Context, status *reimbursement.AuditStatus) error {
	// TODO: 实现创建审核状态逻辑
	return nil
}

// GetAuditStatusByReimbursementID 根据报销单ID获取审核状态
func (r *ReimbursementRepository) GetAuditStatusByReimbursementID(ctx context.Context, reimbursementID string) (*reimbursement.AuditStatus, error) {
	// TODO: 实现获取审核状态逻辑
	return nil, nil
}

// UpdateAuditStatus 更新审核状态
func (r *ReimbursementRepository) UpdateAuditStatus(ctx context.Context, status *reimbursement.AuditStatus) error {
	// TODO: 实现更新审核状态逻辑
	return nil
}

// BeginTx 开始事务
func (r *ReimbursementRepository) BeginTx(ctx context.Context) (reimbursement.Tx, error) {
	// TODO: 实现开始事务逻辑
	return nil, nil
}
