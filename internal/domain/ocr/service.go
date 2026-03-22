// service.go OCR领域服务接口
// 功能点：
// 1. 定义OCR服务接口
// 2. 定义OCR解析服务
// 3. 提供OCR结果验证和转换方法

package ocr

import (
	"context"
	"encoding/json"
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
	parser   InvoiceParser
	repo     Repository
	logger   logger.Logger
	basePath string // 文件存储基础路径
}

// NewParserService 创建OCR解析服务
func NewParserService(parser InvoiceParser, repo Repository, logger logger.Logger, basePath string) *ParserService {
	return &ParserService{
		parser:   parser,
		repo:     repo,
		logger:   logger,
		basePath: basePath,
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

	// 构建完整的文件路径
	fullPath := invoice.ImagePath
	if s.basePath != "" {
		fullPath = s.basePath + "/" + invoice.ImagePath
	}

	// 调用OCR服务解析发票
	ocrResult, err := s.parser.ParseInvoice(ctx, fullPath)
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
	s.updateInvoiceFromOCR(ctx, invoice, ocrResult)
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
func (s *ParserService) updateInvoiceFromOCR(ctx context.Context, invoice *Invoice, ocrResult *InvoiceInfo) {
	invoice.Number = ocrResult.InvoiceNumber

	userType := invoice.Type
	ocrType := ocrResult.InvoiceType

	if userType == "" {
		invoice.Type = ocrType
	}

	if userType != "" && ocrType != "" && userType != ocrType {
		s.logger.WithContext(ctx).Warn("用户填写的发票类型与OCR识别不一致",
			logger.Field{Key: "invoice_id", Value: invoice.ID},
			logger.Field{Key: "user_type", Value: userType},
			logger.Field{Key: "ocr_type", Value: ocrType})

		if invoice.Remarks != "" {
			invoice.Remarks += "; "
		}
		invoice.Remarks += fmt.Sprintf("用户填写类型:%s, OCR识别类型:%s", userType, ocrType)
	}

	if ocrResult.InvoiceDate != "" {
		if parsedDate, err := s.parseDate(ocrResult.InvoiceDate); err == nil {
			invoice.Date = &parsedDate
		}
	}

	invoice.Amount = ocrResult.TotalAmount
	invoice.TaxAmount = ocrResult.TaxAmount

	invoice.BuyerName = ocrResult.BuyerName
	invoice.BuyerTaxNo = ocrResult.BuyerTaxNumber

	invoice.SellerName = ocrResult.SellerName
	invoice.SellerTaxNo = ocrResult.SellerTaxNumber

	invoice.OCRResult = ocrResult.RawText

	invoice.VerificationTime = &ocrResult.ParseTime

	invoice.Type = ocrResult.InvoiceType

	if len(ocrResult.Items) > 0 {
		invoice.TotalItems = len(ocrResult.Items)

		var mainItem *InvoiceItem
		maxAmount := 0.0

		for i := range ocrResult.Items {
			item := &ocrResult.Items[i]
			if item.AmountWithoutTax > maxAmount {
				maxAmount = item.AmountWithoutTax
				mainItem = item
			}
		}

		if mainItem == nil {
			mainItem = &ocrResult.Items[0]
		}

		if mainItem.Name != "" {
			invoice.CommodityName = mainItem.Name
			invoice.MainCommodity = mainItem.Name
		}
		if mainItem.Specification != "" {
			invoice.Specification = mainItem.Specification
		}
		if mainItem.Unit != "" {
			invoice.Unit = mainItem.Unit
		}
		if mainItem.Quantity > 0 {
			invoice.Quantity = mainItem.Quantity
		}
		if mainItem.UnitPrice > 0 {
			invoice.Price = mainItem.UnitPrice
		}

		itemsJSON, err := json.Marshal(ocrResult.Items)
		if err == nil {
			invoice.Items = string(itemsJSON)
			s.logger.WithContext(ctx).Info("存储商品明细",
				logger.Field{Key: "invoice_id", Value: invoice.ID},
				logger.Field{Key: "total_items", Value: invoice.TotalItems},
				logger.Field{Key: "main_commodity", Value: invoice.MainCommodity},
				logger.Field{Key: "main_item_price", Value: mainItem.UnitPrice})
		} else {
			s.logger.WithContext(ctx).Error("序列化商品明细失败",
				logger.Field{Key: "error", Value: err.Error()},
				logger.Field{Key: "invoice_id", Value: invoice.ID})
		}
	}

	if ocrResult.Remarks != "" {
		invoice.Remarks = ocrResult.Remarks
	}

	if ocrResult.IsValid {
		invoice.Status = "recognized"
		invoice.VerificationStatus = "verified"
	} else {
		invoice.Status = "recognition_failed"
		invoice.VerificationStatus = "failed"
		if ocrResult.ErrorMessage != "" {
			if invoice.Remarks != "" {
				invoice.Remarks += "; " + ocrResult.ErrorMessage
			} else {
				invoice.Remarks = ocrResult.ErrorMessage
			}
		}
	}

	if ocrResult.InvoiceType != "" {
		invoice.IsVAT = strings.Contains(ocrResult.InvoiceType, "增值税")
	}

	s.inferExtendedFields(invoice, ocrResult)
}

// inferExtendedFields 智能推断扩展字段
func (s *ParserService) inferExtendedFields(invoice *Invoice, ocrResult *InvoiceInfo) {
	s.inferVATRate(invoice, ocrResult)
	s.inferInvoiceCategory(invoice, ocrResult)
	s.inferMerchantType(invoice, ocrResult)
	s.inferInvoiceDescription(invoice, ocrResult)
	s.inferIsElectronic(invoice, ocrResult)
}

// inferVATRate 推断增值税率
func (s *ParserService) inferVATRate(invoice *Invoice, ocrResult *InvoiceInfo) {
	if invoice.Amount > 0 && invoice.TaxAmount > 0 {
		invoice.VATRate = (invoice.TaxAmount / invoice.Amount) * 100
	}
}

// inferInvoiceCategory 推断发票类别
func (s *ParserService) inferInvoiceCategory(invoice *Invoice, ocrResult *InvoiceInfo) {
	invoiceType := invoice.Category
	if invoiceType == "" {
		invoiceType = ocrResult.InvoiceType
	}

	sellerName := ocrResult.SellerName
	mainCommodity := invoice.MainCommodity

	switch {
	case strings.Contains(invoiceType, "住宿") || strings.Contains(sellerName, "酒店") || strings.Contains(sellerName, "宾馆") || strings.Contains(mainCommodity, "住宿"):
		invoice.Category = "差旅费"
		invoice.SubCategory = "住宿费"
	case strings.Contains(invoiceType, "交通") || strings.Contains(sellerName, "航空") || strings.Contains(sellerName, "铁路") || strings.Contains(sellerName, "客运") || strings.Contains(mainCommodity, "机票") || strings.Contains(mainCommodity, "火车票"):
		invoice.Category = "差旅费"
		invoice.SubCategory = "交通费"
	case strings.Contains(invoiceType, "餐饮") || strings.Contains(sellerName, "餐厅") || strings.Contains(sellerName, "饭店") || strings.Contains(sellerName, "美食") || strings.Contains(mainCommodity, "餐饮"):
		invoice.Category = "招待费"
		invoice.SubCategory = "餐饮费"
	case strings.Contains(invoiceType, "办公用品") || strings.Contains(sellerName, "文具") || strings.Contains(sellerName, "办公") || strings.Contains(mainCommodity, "办公用品"):
		invoice.Category = "办公费"
		invoice.SubCategory = "办公用品"
	case strings.Contains(invoiceType, "培训") || strings.Contains(sellerName, "培训") || strings.Contains(sellerName, "教育") || strings.Contains(mainCommodity, "培训"):
		invoice.Category = "培训费"
		invoice.SubCategory = "培训费"
	case strings.Contains(invoiceType, "会议") || strings.Contains(sellerName, "会议") || strings.Contains(mainCommodity, "会议"):
		invoice.Category = "会议费"
		invoice.SubCategory = "会议费"
	case strings.Contains(invoiceType, "通信") || strings.Contains(sellerName, "通信") || strings.Contains(sellerName, "电信") || strings.Contains(mainCommodity, "话费"):
		invoice.Category = "通信费"
		invoice.SubCategory = "通信费"
	case strings.Contains(invoiceType, "租赁") || strings.Contains(sellerName, "租赁") || strings.Contains(mainCommodity, "租赁"):
		invoice.Category = "租赁费"
		invoice.SubCategory = "租赁费"
	case strings.Contains(invoiceType, "服务") || strings.Contains(sellerName, "服务") || strings.Contains(mainCommodity, "服务"):
		invoice.Category = "服务费"
		invoice.SubCategory = "服务费"
	case strings.Contains(invoiceType, "咨询") || strings.Contains(sellerName, "咨询") || strings.Contains(mainCommodity, "咨询"):
		invoice.Category = "咨询费"
		invoice.SubCategory = "咨询费"
	default:
		invoice.Category = "其他费用"
		invoice.SubCategory = "其他"
	}
}

// inferMerchantType 推断商户类型
func (s *ParserService) inferMerchantType(invoice *Invoice, ocrResult *InvoiceInfo) {
	sellerName := ocrResult.SellerName

	switch {
	case strings.Contains(sellerName, "酒店") || strings.Contains(sellerName, "宾馆") || strings.Contains(sellerName, "旅馆"):
		invoice.MerchantType = "酒店"
	case strings.Contains(sellerName, "餐厅") || strings.Contains(sellerName, "饭店") || strings.Contains(sellerName, "美食"):
		invoice.MerchantType = "餐厅"
	case strings.Contains(sellerName, "航空") || strings.Contains(sellerName, "机场"):
		invoice.MerchantType = "航空公司"
	case strings.Contains(sellerName, "铁路") || strings.Contains(sellerName, "高铁"):
		invoice.MerchantType = "铁路公司"
	case strings.Contains(sellerName, "超市") || strings.Contains(sellerName, "商场"):
		invoice.MerchantType = "超市"
	case strings.Contains(sellerName, "加油站") || strings.Contains(sellerName, "石油"):
		invoice.MerchantType = "加油站"
	case strings.Contains(sellerName, "医院") || strings.Contains(sellerName, "诊所"):
		invoice.MerchantType = "医疗机构"
	case strings.Contains(sellerName, "银行"):
		invoice.MerchantType = "银行"
	}
}

// inferInvoiceDescription 推断发票描述
func (s *ParserService) inferInvoiceDescription(invoice *Invoice, ocrResult *InvoiceInfo) {
	var descParts []string

	if ocrResult.InvoiceName != "" {
		descParts = append(descParts, ocrResult.InvoiceName)
	}

	if ocrResult.SellerName != "" {
		descParts = append(descParts, "销售方: "+ocrResult.SellerName)
	}

	if ocrResult.InvoiceType != "" {
		descParts = append(descParts, "类型: "+ocrResult.InvoiceType)
	}

	if invoice.Amount > 0 {
		descParts = append(descParts, fmt.Sprintf("金额: %.2f", invoice.Amount))
	}

	invoice.Description = strings.Join(descParts, "; ")
}

// inferIsElectronic 推断是否为电子发票
func (s *ParserService) inferIsElectronic(invoice *Invoice, ocrResult *InvoiceInfo) {
	invoiceType := ocrResult.InvoiceType
	invoice.IsElectronic = strings.Contains(invoiceType, "电子") || strings.Contains(invoiceType, "数电")
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
		"2006年01月02日",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析日期格式: %s", dateStr)
}
