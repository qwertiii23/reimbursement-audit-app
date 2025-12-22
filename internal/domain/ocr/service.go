// service.go OCR领域服务接口
// 功能点：
// 1. 定义OCR服务接口
// 2. 定义OCR解析服务
// 3. 提供OCR结果验证和转换方法

package ocr

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"reimbursement-audit/internal/pkg/logger"
)

// InvoiceParser 发票解析器接口
type InvoiceParser interface {
	// ParseInvoice 解析发票图片，返回发票信息
	ParseInvoice(ctx context.Context, imagePath string) (*InvoiceInfo, error)
}

// ParserService OCR解析领域服务
type ParserService struct {
	parser InvoiceParser
	repo   Repository
	logger logger.Logger
}

// NewParserService 创建OCR解析服务
func NewParserService(parser InvoiceParser, repo Repository, logger logger.Logger) *ParserService {
	return &ParserService{
		parser: parser,
		repo:   repo,
		logger: logger,
	}
}

// ParseInvoiceImage 解析发票图片并更新数据库
func (s *ParserService) ParseInvoiceImage(ctx context.Context, invoiceID string) error {
	// 从数据库获取发票信息
	invoice, err := s.repo.GetInvoiceByID(ctx, invoiceID)
	if err != nil {
		s.logger.WithContext(ctx).Error("获取发票信息失败",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "invoice_id", Value: invoiceID})
		return fmt.Errorf("获取发票信息失败: %w", err)
	}

	s.logger.WithContext(ctx).Info("开始解析发票图片",
		logger.Field{Key: "invoice_id", Value: invoiceID},
		logger.Field{Key: "image_path", Value: invoice.ImagePath})

	// 调用OCR服务解析发票
	ocrResult, err := s.parser.ParseInvoice(ctx, invoice.ImagePath)
	if err != nil {
		s.logger.WithContext(ctx).Error("OCR解析失败",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "invoice_id", Value: invoiceID},
			logger.Field{Key: "image_path", Value: invoice.ImagePath})

		// 更新发票状态为解析失败
		invoice.Status = "解析失败"
		invoice.UpdatedAt = time.Now()
		if updateErr := s.repo.UpdateInvoice(ctx, invoice); updateErr != nil {
			s.logger.WithContext(ctx).Error("更新发票状态失败",
				logger.Field{Key: "error", Value: updateErr.Error()},
				logger.Field{Key: "invoice_id", Value: invoiceID})
		}

		return fmt.Errorf("OCR解析失败: %w", err)
	}

	// 验证OCR解析结果
	isValid, errMsg := ocrResult.Validate()
	if !isValid {
		s.logger.WithContext(ctx).Warn("OCR解析结果验证失败",
			logger.Field{Key: "error", Value: errMsg},
			logger.Field{Key: "invoice_id", Value: invoiceID})

		// 更新发票状态为无效
		invoice.Status = "无效"
		invoice.UpdatedAt = time.Now()
		if updateErr := s.repo.UpdateInvoice(ctx, invoice); updateErr != nil {
			s.logger.WithContext(ctx).Error("更新发票状态失败",
				logger.Field{Key: "error", Value: updateErr.Error()},
				logger.Field{Key: "invoice_id", Value: invoiceID})
		}

		return fmt.Errorf("OCR解析结果验证失败: %s", errMsg)
	}

	// 更新发票信息
	s.updateInvoiceFromOCR(invoice, ocrResult)
	invoice.Status = "已识别"
	invoice.UpdatedAt = time.Now()

	// 保存更新后的发票信息
	if err := s.repo.UpdateInvoice(ctx, invoice); err != nil {
		s.logger.WithContext(ctx).Error("更新发票信息失败",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "invoice_id", Value: invoiceID})
		return fmt.Errorf("更新发票信息失败: %w", err)
	}

	s.logger.WithContext(ctx).Info("发票解析完成",
		logger.Field{Key: "invoice_id", Value: invoiceID},
		logger.Field{Key: "invoice_code", Value: invoice.Code},
		logger.Field{Key: "invoice_number", Value: invoice.Number},
		logger.Field{Key: "amount", Value: invoice.Amount})

	return nil
}

// ParseInvoice 解析发票图片，实现InvoiceParser接口
func (s *ParserService) ParseInvoice(ctx context.Context, imagePath string) (*InvoiceInfo, error) {
	return s.parser.ParseInvoice(ctx, imagePath)
}

// ParseInvoicesByReimbursementID 根据报销单ID解析所有关联的发票
func (s *ParserService) ParseInvoicesByReimbursementID(ctx context.Context, reimbursementID string) error {
	// 获取报销单的所有发票
	invoices, err := s.repo.ListInvoicesByReimbursementID(ctx, reimbursementID)
	if err != nil {
		s.logger.WithContext(ctx).Error("获取报销单发票列表失败",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "reimbursement_id", Value: reimbursementID})
		return fmt.Errorf("获取报销单发票列表失败: %w", err)
	}

	s.logger.WithContext(ctx).Info("开始批量解析发票",
		logger.Field{Key: "reimbursement_id", Value: reimbursementID},
		logger.Field{Key: "invoice_count", Value: len(invoices)})

	// 逐个解析发票
	var errors []string
	for _, invoice := range invoices {
		if err := s.ParseInvoiceImage(ctx, invoice.ID); err != nil {
			errors = append(errors, fmt.Sprintf("发票 %s 解析失败: %s", invoice.ID, err.Error()))
		}
	}

	if len(errors) > 0 {
		s.logger.WithContext(ctx).Error("批量解析发票完成，部分失败",
			logger.Field{Key: "reimbursement_id", Value: reimbursementID},
			logger.Field{Key: "success_count", Value: len(invoices) - len(errors)},
			logger.Field{Key: "failure_count", Value: len(errors)},
			logger.Field{Key: "errors", Value: strings.Join(errors, "; ")})
		return fmt.Errorf("批量解析完成，%d个失败: %s", len(errors), strings.Join(errors, "; "))
	}

	s.logger.WithContext(ctx).Info("批量解析发票全部成功",
		logger.Field{Key: "reimbursement_id", Value: reimbursementID},
		logger.Field{Key: "invoice_count", Value: len(invoices)})

	return nil
}

// updateInvoiceFromOCR 使用OCR结果更新发票信息
func (s *ParserService) updateInvoiceFromOCR(invoice *Invoice, ocrResult *InvoiceInfo) {
	// 更新发票基本信息
	invoice.Code = ocrResult.InvoiceCode
	invoice.Number = ocrResult.InvoiceNumber
	invoice.Type = ocrResult.InvoiceType

	// 解析日期字符串为time.Time
	if ocrResult.InvoiceDate != "" {
		if parsedDate, err := s.parseDate(ocrResult.InvoiceDate); err == nil {
			invoice.Date = parsedDate
		}
	}

	// 更新金额信息
	invoice.Amount = ocrResult.TotalAmount
	invoice.TaxAmount = ocrResult.TaxAmount

	// 更新购方信息
	invoice.BuyerName = ocrResult.BuyerName
	invoice.BuyerTaxNo = ocrResult.BuyerTaxNumber

	// 更新销方信息
	invoice.SellerName = ocrResult.SellerName
	invoice.SellerTaxNo = ocrResult.SellerTaxNumber

	// 更新OCR识别结果
	invoice.OCRResult = ocrResult.RawText
}

// parseDate 解析日期字符串为time.Time
func (s *ParserService) parseDate(dateStr string) (time.Time, error) {
	// 尝试YYYYMMDD格式
	if len(dateStr) == 8 {
		if year, err := strconv.Atoi(dateStr[:4]); err == nil {
			if month, err := strconv.Atoi(dateStr[4:6]); err == nil && month >= 1 && month <= 12 {
				if day, err := strconv.Atoi(dateStr[6:8]); err == nil && day >= 1 && day <= 31 {
					return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
				}
			}
		}
	}

	// 尝试YYYY-MM-DD格式
	if len(dateStr) == 10 && dateStr[4] == '-' && dateStr[7] == '-' {
		if year, err := strconv.Atoi(dateStr[:4]); err == nil {
			if month, err := strconv.Atoi(dateStr[5:7]); err == nil && month >= 1 && month <= 12 {
				if day, err := strconv.Atoi(dateStr[8:10]); err == nil && day >= 1 && day <= 31 {
					return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
				}
			}
		}
	}

	// 尝试其他常见格式
	formats := []string{
		"20060102",
		"2006-01-02",
		"2006/01/02",
		"2006.01.02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析日期格式: %s", dateStr)
}
