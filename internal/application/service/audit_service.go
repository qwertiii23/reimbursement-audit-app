package service

import (
	"context"
	"fmt"

	"reimbursement-audit/internal/api/request"
	"reimbursement-audit/internal/api/response"
	"reimbursement-audit/internal/domain/audit"
	"reimbursement-audit/internal/pkg/logger"
)

// AuditApplicationService 审核应用服务
type AuditApplicationService struct {
	auditService *audit.Service
	logger       logger.Logger
}

// NewAuditApplicationService 创建审核应用服务
func NewAuditApplicationService(
	auditService *audit.Service,
	logger logger.Logger,
) *AuditApplicationService {
	return &AuditApplicationService{
		auditService: auditService,
		logger:       logger,
	}
}

// StartAudit 开始审核用例
func (s *AuditApplicationService) StartAudit(ctx context.Context, req *request.StartAuditRequest) (*response.AuditResponse, error) {
	s.logger.WithContext(ctx).Info("开始审核用例", logger.NewField("reimbursement_id", req.ReimbursementID))

	auditResult, err := s.auditService.StartAudit(ctx, req.ReimbursementID)
	if err != nil {
		s.logger.WithContext(ctx).Error("审核失败", logger.NewField("error", err))
		return nil, fmt.Errorf("审核失败: %w", err)
	}

	return response.NewAuditResponse(auditResult), nil
}

// GetAuditStatus 获取审核状态用例
func (s *AuditApplicationService) GetAuditStatus(ctx context.Context, auditID string) (*response.AuditStatusResponse, error) {
	s.logger.WithContext(ctx).Info("获取审核状态", logger.NewField("audit_id", auditID))

	auditResult, err := s.auditService.GetAuditStatus(ctx, auditID)
	if err != nil {
		s.logger.WithContext(ctx).Error("获取审核状态失败", logger.NewField("error", err))
		return nil, fmt.Errorf("获取审核状态失败: %w", err)
	}

	return response.NewAuditStatusResponse(auditResult), nil
}

// GetAuditResult 获取审核结果用例
func (s *AuditApplicationService) GetAuditResult(ctx context.Context, auditID string) (*response.AuditResultResponse, error) {
	s.logger.WithContext(ctx).Info("获取审核结果", logger.NewField("audit_id", auditID))

	auditResult, err := s.auditService.GetAuditStatus(ctx, auditID)
	if err != nil {
		s.logger.WithContext(ctx).Error("获取审核结果失败", logger.NewField("error", err))
		return nil, fmt.Errorf("获取审核结果失败: %w", err)
	}

	return response.NewAuditResultResponse(auditResult), nil
}

// GetAuditByReimbursementID 根据报销单ID获取审核结果用例
func (s *AuditApplicationService) GetAuditByReimbursementID(ctx context.Context, reimbursementID string) (*response.AuditResultResponse, error) {
	s.logger.WithContext(ctx).Info("根据报销单ID获取审核结果", logger.NewField("reimbursement_id", reimbursementID))

	auditResult, err := s.auditService.GetAuditByReimbursementID(ctx, reimbursementID)
	if err != nil {
		s.logger.WithContext(ctx).Error("获取审核结果失败", logger.NewField("error", err))
		return nil, fmt.Errorf("获取审核结果失败: %w", err)
	}

	return response.NewAuditResultResponse(auditResult), nil
}

// RetryAudit 重试审核用例
func (s *AuditApplicationService) RetryAudit(ctx context.Context, auditID string) (*response.AuditResponse, error) {
	s.logger.WithContext(ctx).Info("重试审核", logger.NewField("audit_id", auditID))

	auditResult, err := s.auditService.RetryAudit(ctx, auditID)
	if err != nil {
		s.logger.WithContext(ctx).Error("重试审核失败", logger.NewField("error", err))
		return nil, fmt.Errorf("重试审核失败: %w", err)
	}

	return response.NewAuditResponse(auditResult), nil
}
