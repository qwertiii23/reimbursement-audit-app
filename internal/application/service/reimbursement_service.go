// reimbursement_service.go 报销单应用服务
// 功能点：
// 1. 处理报销单业务流程编排
// 2. 协调领域服务和基础设施
// 3. 处理事务边界
// 4. 提供用例级别的接口

package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"reimbursement-audit/internal/api/request"
	"reimbursement-audit/internal/api/response"
	"reimbursement-audit/internal/domain/ocr"
	"reimbursement-audit/internal/domain/reimbursement"
	storage "reimbursement-audit/internal/infra/storage/file"
	"reimbursement-audit/internal/pkg/logger"

	"github.com/google/uuid"
)

// ReimbursementApplicationService 报销单应用服务
type ReimbursementApplicationService struct {
	reimbursementRepo    reimbursement.Repository
	reimbursementService reimbursement.Service
	ocrService           ocr.InvoiceParser
	ocrRepo              ocr.Repository
	fileService          *storage.Service
	logger               logger.Logger
}

// NewReimbursementApplicationService 创建报销单应用服务
func NewReimbursementApplicationService(
	reimbursementRepo reimbursement.Repository,
	reimbursementService reimbursement.Service,
	ocrService ocr.InvoiceParser,
	ocrRepo ocr.Repository,
	fileService *storage.Service,
	logger logger.Logger,
) *ReimbursementApplicationService {
	return &ReimbursementApplicationService{
		reimbursementRepo:    reimbursementRepo,
		reimbursementService: reimbursementService,
		ocrService:           ocrService,
		ocrRepo:              ocrRepo,
		fileService:          fileService,
		logger:               logger,
	}
}

// CreateReimbursement 创建报销单用例
func (s *ReimbursementApplicationService) CreateReimbursement(ctx context.Context, req *request.ReimbursementUploadRequest) (*response.ReimbursementUploadResponse, error) {
	// 清理和标准化请求数据
	req.Sanitize()

	// 校验请求数据
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("参数校验失败: %w", err)
	}

	// 创建领域服务请求
	domainReq := &reimbursement.CreateReimbursementRequest{
		UserID:      req.UserID,
		UserName:    req.UserName,
		Department:  req.Department,
		Category:    req.Category,
		Reason:      req.Reason,
		Description: req.Description,
		TotalAmount: req.TotalAmount,
		ApplyDate:   req.ApplyDate,
		ExpenseDate: req.ExpenseDate,
	}

	// 调用领域服务创建报销单
	reimbursementModel, err := s.reimbursementService.CreateReimbursement(ctx, domainReq)
	if err != nil {
		return nil, fmt.Errorf("创建报销单失败: %w", err)
	}

	// 创建响应数据
	return response.NewReimbursementUploadResponse(
		reimbursementModel.ID,
		reimbursementModel.UserID,
		reimbursementModel.UserName,
		reimbursementModel.Type,
		reimbursementModel.TotalAmount,
		reimbursementModel.Status,
		reimbursementModel.CreatedAt,
	), nil
}

// UploadInvoice 上传发票用例
func (s *ReimbursementApplicationService) UploadInvoice(ctx context.Context, reimbursementID string, fileHeader *multipart.FileHeader) (*response.InvoiceUploadResponse, error) {
	// 验证报销单是否存在
	_, err := s.reimbursementRepo.GetReimbursementByID(ctx, reimbursementID)
	if err != nil {
		return nil, fmt.Errorf("报销单不存在: %w", err)
	}

	// 上传发票文件到存储服务
	fileInfo, err := s.fileService.UploadInvoice(ctx, fileHeader)
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	// 创建发票记录
	now := time.Now()
	invoice := &ocr.Invoice{
		ID:              uuid.New().String(),
		ReimbursementID: reimbursementID,
		ImagePath:       fileInfo.Path,
		Status:          "待识别", // 初始状态为待识别，等待OCR处理
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// 保存发票记录到数据库
	if err := s.ocrRepo.CreateInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("保存发票记录失败: %w", err)
	}

	// 异步进行OCR解析
	go s.processOCRAsync(ctx, invoice.ID)

	// 创建响应数据
	return response.NewInvoiceUploadResponse(
		invoice.ID,
		reimbursementID,
		fileInfo.Path,
		fileInfo.Size,
		invoice.Status,
	), nil
}

// BatchUploadInvoices 批量上传发票用例
func (s *ReimbursementApplicationService) BatchUploadInvoices(ctx context.Context, reimbursementID string, fileHeaders []interface{}) (*response.BatchUploadResponse, error) {
	// 验证报销单是否存在
	_, err := s.reimbursementRepo.GetReimbursementByID(ctx, reimbursementID)
	if err != nil {
		return nil, fmt.Errorf("报销单不存在: %w", err)
	}

	// 限制批量上传数量
	maxBatchSize := 10
	if len(fileHeaders) > maxBatchSize {
		return nil, fmt.Errorf("批量上传文件数量不能超过%d个", maxBatchSize)
	}

	// 生成批次ID
	batchID := uuid.New().String()

	// 存储成功上传的发票信息
	var successfulInvoices []*ocr.Invoice
	var invoiceResponses []response.InvoiceUploadResponse
	var errors []string

	// 逐个处理文件上传
	for _, fileHeader := range fileHeaders {
		// 类型断言
		multipartFileHeader, ok := fileHeader.(*multipart.FileHeader)
		if !ok {
			errors = append(errors, "文件类型错误")
			continue
		}

		// 上传文件
		fileInfo, err := s.fileService.UploadInvoice(ctx, multipartFileHeader)
		if err != nil {
			errors = append(errors, fmt.Sprintf("上传文件失败: %s", err.Error()))
			continue
		}

		// 创建发票记录
		now := time.Now()
		invoice := &ocr.Invoice{
			ID:              uuid.New().String(),
			ReimbursementID: reimbursementID,
			ImagePath:       fileInfo.Path,
			Status:          "待识别", // 初始状态为待识别，等待OCR处理
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		// 保存发票记录到数据库
		if err := s.ocrRepo.CreateInvoice(ctx, invoice); err != nil {
			errors = append(errors, fmt.Sprintf("保存发票记录失败: %s", err.Error()))
			continue
		}

		successfulInvoices = append(successfulInvoices, invoice)
		invoiceResponses = append(invoiceResponses, *response.NewInvoiceUploadResponse(
			invoice.ID,
			reimbursementID,
			fileInfo.Path,
			fileInfo.Size,
			invoice.Status,
		))
	}

	// 异步进行批量OCR解析
	go s.processBatchOCRAsync(ctx, successfulInvoices)

	// 创建批量上传响应
	batchResponse := response.NewBatchUploadResponse(
		batchID,
		len(fileHeaders),
		len(successfulInvoices),
		len(errors),
	)

	// 设置响应数据
	batchResponse.Invoices = invoiceResponses
	batchResponse.Errors = errors

	return batchResponse, nil
}

// GetReimbursementDetail 获取报销单详情（包括发票列表）
func (s *ReimbursementApplicationService) GetReimbursementDetail(ctx context.Context, id string) (*reimbursement.Reimbursement, error) {
	// 获取报销单基本信息
	reimb, err := s.reimbursementRepo.GetReimbursementByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取报销单失败: %w", err)
	}

	// 获取关联的发票列表
	invoices, err := s.ocrRepo.ListInvoicesByReimbursementID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取发票列表失败: %w", err)
	}

	// 组装完整信息
	reimb.Invoices = invoices

	return reimb, nil
}

// processOCRAsync 异步处理OCR解析
func (s *ReimbursementApplicationService) processOCRAsync(ctx context.Context, invoiceID string) {
	if s.ocrService == nil {
		s.logger.WithContext(ctx).Warn("OCR服务未配置", logger.NewField("invoice_id", invoiceID))
		return
	}

	// 调用OCR解析服务
	if parserService, ok := s.ocrService.(*ocr.ParserService); ok {
		if err := parserService.ParseInvoiceImage(ctx, invoiceID); err != nil {
			// 记录错误日志
			s.logger.WithContext(ctx).Error("OCR解析失败",
				logger.NewField("invoice_id", invoiceID),
				logger.NewField("error", err.Error()))
		}
	} else {
		// 如果不是ParserService，则无法通过invoiceID解析
		s.logger.WithContext(ctx).Error("OCR解析失败, 不支持的解析服务类型",
			logger.NewField("invoice_id", invoiceID))
	}
}

// processBatchOCRAsync 异步处理批量OCR解析
func (s *ReimbursementApplicationService) processBatchOCRAsync(ctx context.Context, invoices []*ocr.Invoice) {
	if s.ocrService == nil {
		s.logger.WithContext(ctx).Warn("OCR服务未配置", logger.NewField("batch_size", len(invoices)))
		return
	}

	// 逐个调用OCR解析服务
	if parserService, ok := s.ocrService.(*ocr.ParserService); ok {
		for _, invoice := range invoices {
			if err := parserService.ParseInvoiceImage(ctx, invoice.ID); err != nil {
				// 记录错误日志
				s.logger.WithContext(ctx).Error("OCR解析失败",
					logger.NewField("invoice_id", invoice.ID),
					logger.NewField("error", err.Error()))
			}
		}
	} else {
		// 如果不是ParserService，则无法通过invoiceID解析
		for _, invoice := range invoices {
			s.logger.WithContext(ctx).Error("OCR解析失败, 不支持的解析服务类型",
				logger.NewField("invoice_id", invoice.ID))
		}
	}
}
