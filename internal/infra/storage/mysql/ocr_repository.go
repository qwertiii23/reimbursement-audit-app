package mysql

import (
	"context"
	"errors"

	"reimbursement-audit/internal/domain/ocr"
	"reimbursement-audit/internal/pkg/logger"

	"gorm.io/gorm"
)

// OCRRepository OCR仓储实现
type OCRRepository struct {
	client *Client
	logger logger.Logger
}

// NewOCRRepository 创建OCR仓储实例
func NewOCRRepository (client *Client, logger logger.Logger) ocr.Repository {
	return &OCRRepository{client: client, logger: logger}
}

// CreateInvoice 创建发票
func (r *OCRRepository) CreateInvoice(ctx context.Context, invoice *ocr.Invoice) error {
	// 使用GORM创建发票记录
	result := r.client.GetDB().WithContext(ctx).Create(invoice)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("创建发票失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("invoice_id", invoice.ID),
			logger.NewField("reimbursement_id", invoice.ReimbursementID))
		return result.Error
	}

	return nil
}

// CreateInvoices 批量创建发票
func (r *OCRRepository) CreateInvoices(ctx context.Context, invoices []*ocr.Invoice) error {
	if len(invoices) == 0 {
		return nil
	}

	// 使用GORM批量创建发票记录
	result := r.client.GetDB().WithContext(ctx).CreateInBatches(invoices, 100) // 每批最多100条
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("批量创建发票失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("batch_size", len(invoices)))
		return result.Error
	}

	return nil
}

// GetInvoiceByID 根据ID获取发票
func (r *OCRRepository) GetInvoiceByID(ctx context.Context, id string) (*ocr.Invoice, error) {
	var invoice ocr.Invoice

	// 使用GORM查询发票
	result := r.client.GetDB().WithContext(ctx).Where("id = ?", id).First(&invoice)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("发票不存在",
				logger.NewField("invoice_id", id))
			return nil, result.Error
		}
		r.logger.WithContext(ctx).Error("查询发票失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("invoice_id", id))
		return nil, result.Error
	}

	return &invoice, nil
}

// UpdateInvoice 更新发票
func (r *OCRRepository) UpdateInvoice(ctx context.Context, invoice *ocr.Invoice) error {
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
		r.logger.WithContext(ctx).Error("更新发票失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("invoice_id", invoice.ID))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("发票不存在，更新失败",
			logger.NewField("invoice_id", invoice.ID))
		return result.Error
	}

	return nil
}

// DeleteInvoice 删除发票
func (r *OCRRepository) DeleteInvoice(ctx context.Context, id string) error {
	// 使用GORM删除发票
	result := r.client.GetDB().WithContext(ctx).Where("id = ?", id).Delete(&ocr.Invoice{})

	if result.Error != nil {
		r.logger.WithContext(ctx).Error("删除发票失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("invoice_id", id))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("发票不存在，删除失败",
			logger.NewField("invoice_id", id))
		return result.Error
	}

	return nil
}

// ListInvoicesByReimbursementID 根据报销单ID获取发票列表
func (r *OCRRepository) ListInvoicesByReimbursementID(ctx context.Context, reimbursementID string) ([]*ocr.Invoice, error) {
	var invoices []*ocr.Invoice

	// 使用GORM查询发票列表
	result := r.client.GetDB().WithContext(ctx).
		Where("reimbursement_id = ?", reimbursementID).
		Order("created_at ASC").
		Find(&invoices)

	if result.Error != nil {
		r.logger.WithContext(ctx).Error("获取发票列表失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("reimbursement_id", reimbursementID))
		return nil, result.Error
	}

	return invoices, nil
}
