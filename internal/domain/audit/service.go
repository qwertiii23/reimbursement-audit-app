package audit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"reimbursement-audit/internal/domain/rag"
	"reimbursement-audit/internal/domain/reimbursement"
	ruleenginedomain "reimbursement-audit/internal/domain/ruleengine"
	"reimbursement-audit/internal/pkg/logger"

	"github.com/google/uuid"
)

// Service 审核服务
type Service struct {
	repo              Repository
	reimbursementRepo reimbursement.Repository
	ruleEngineService *ruleenginedomain.RuleEngineService
	ragService        *rag.RAGService
	logger            logger.Logger
}

// NewService 创建审核服务
func NewService(
	repo Repository,
	reimbursementRepo reimbursement.Repository,
	ruleEngineService *ruleenginedomain.RuleEngineService,
	ragService *rag.RAGService,
	logger logger.Logger,
) *Service {
	return &Service{
		repo:              repo,
		reimbursementRepo: reimbursementRepo,
		ruleEngineService: ruleEngineService,
		ragService:        ragService,
		logger:            logger,
	}
}

// StartAudit 开始审核
func (s *Service) StartAudit(ctx context.Context, reimbursementID string) (*AuditResult, error) {
	startTime := time.Now()

	s.logger.WithContext(ctx).Info("开始审核", logger.NewField("reimbursement_id", reimbursementID))

	reimbursement, err := s.reimbursementRepo.GetReimbursementByID(ctx, reimbursementID)
	reimbursement.Invoices, err = s.reimbursementRepo.GetInvoicesByReimbursementID(ctx, reimbursementID)
	if err != nil {
		s.logger.WithContext(ctx).Error("获取报销单失败", logger.NewField("error", err))
		return nil, fmt.Errorf("获取报销单失败: %w", err)
	}

	audit := &AuditResult{
		ID:              uuid.New().String(),
		ReimbursementID: reimbursementID,
		Status:          AuditStatusRunning,
		WorkflowStatus:  WorkflowStatusSubmitted,
		StartedAt:       startTime,
		CreatedAt:       startTime,
		UpdatedAt:       startTime,
	}

	if err := s.repo.CreateAudit(ctx, audit); err != nil {
		s.logger.WithContext(ctx).Error("创建审核记录失败", logger.NewField("error", err))
		return nil, fmt.Errorf("创建审核记录失败: %w", err)
	}

	reimbursement.Status = "auditing"
	reimbursement.AuditID = audit.ID
	if err := s.reimbursementRepo.UpdateReimbursement(ctx, reimbursement); err != nil {
		s.logger.WithContext(ctx).Error("更新报销单状态失败", logger.NewField("error", err))
		return nil, fmt.Errorf("更新报销单状态失败: %w", err)
	}

	startFlowLog := &AuditFlowLog{
		ID:              uuid.New().String(),
		ReimbursementID: reimbursementID,
		AuditID:         audit.ID,
		FlowStatus:      FlowStatusIntelligentStart,
		FlowType:        FlowTypeIntelligent,
		Action:          FlowActionStartAudit,
		CreatedAt:       startTime,
	}

	if err := s.repo.CreateFlowLog(ctx, startFlowLog); err != nil {
		s.logger.WithContext(ctx).Error("创建流程日志失败", logger.NewField("error", err))
	}

	audit.WorkflowStatus = WorkflowStatusRuleAudit
	audit.UpdatedAt = startTime
	s.repo.UpdateAudit(ctx, audit)

	ruleResults, err := s.executeRuleValidation(ctx, reimbursement)
	if err != nil {
		s.logger.WithContext(ctx).Error("规则校验失败", logger.NewField("error", err))
		audit.Status = AuditStatusCompleted
		audit.WorkflowStatus = WorkflowStatusRuleFailed
		audit.Reason = fmt.Sprintf("规则校验失败: %s", err.Error())
		audit.FinalPass = false
		audit.RulePass = false
		audit.RAGPass = false
		audit.RiskScore = 1.0
		audit.RiskLevel = "高风险"
		completedTime := time.Now()
		audit.CompletedAt = &completedTime
		audit.Duration = completedTime.Sub(startTime).Milliseconds()
		audit.UpdatedAt = completedTime
		s.repo.UpdateAudit(ctx, audit)

		failFlowLog := &AuditFlowLog{
			ID:              uuid.New().String(),
			ReimbursementID: reimbursementID,
			AuditID:         audit.ID,
			FlowStatus:      FlowStatusIntelligentFail,
			FlowType:        FlowTypeIntelligent,
			Action:          FlowActionRejectAudit,
			Reason:          &audit.Reason,
			CreatedAt:       completedTime,
		}
		s.repo.CreateFlowLog(ctx, failFlowLog)

		return audit, nil
	}

	audit.RuleResults = ruleResults
	rulePass := s.checkRulePass(ruleResults)
	audit.RulePass = rulePass

	if !rulePass {
		audit.WorkflowStatus = WorkflowStatusRuleFailed
		audit.UpdatedAt = time.Now()
		s.repo.UpdateAudit(ctx, audit)

		reimbursementInfo := s.buildReimbursementInfo(reimbursement)

		audit.WorkflowStatus = WorkflowStatusRAGAudit
		audit.UpdatedAt = time.Now()
		s.repo.UpdateAudit(ctx, audit)

		ragResult, err := s.executeRAGAnalysis(ctx, reimbursementInfo)
		if err != nil {
			s.logger.WithContext(ctx).Error("RAG分析失败", logger.NewField("error", err))
			audit.RAGPass = false
		} else {
			audit.RAGResults = ragResult
			audit.RAGPass = ragResult != nil && ragResult.Confidence > 0.6
		}

		audit.FinalPass = false
		audit.Reason = "规则校验未通过"
		audit.RiskScore = s.calculateRiskScore(audit)
		audit.RiskLevel = s.determineRiskLevel(audit.RiskScore)
		audit.Suggestions = s.generateSuggestions(audit)
		completedTime := time.Now()
		audit.CompletedAt = &completedTime
		audit.Duration = completedTime.Sub(startTime).Milliseconds()
		audit.UpdatedAt = completedTime

		if err := s.repo.UpdateAudit(ctx, audit); err != nil {
			s.logger.WithContext(ctx).Error("更新审核记录失败", logger.NewField("error", err))
			return nil, fmt.Errorf("更新审核记录失败: %w", err)
		}

		reimbursement.Status = "rejected"
		if err := s.reimbursementRepo.UpdateReimbursement(ctx, reimbursement); err != nil {
			s.logger.WithContext(ctx).Error("更新报销单状态失败", logger.NewField("error", err))
		}

		failFlowLog := &AuditFlowLog{
			ID:              uuid.New().String(),
			ReimbursementID: reimbursementID,
			AuditID:         audit.ID,
			FlowStatus:      FlowStatusIntelligentFail,
			FlowType:        FlowTypeIntelligent,
			Action:          FlowActionRejectAudit,
			Reason:          &audit.Reason,
			CreatedAt:       completedTime,
		}
		if err := s.repo.CreateFlowLog(ctx, failFlowLog); err != nil {
			s.logger.WithContext(ctx).Error("创建流程日志失败", logger.NewField("error", err))
		}

		return audit, nil
	}

	audit.WorkflowStatus = WorkflowStatusRulePassed
	audit.UpdatedAt = time.Now()
	s.repo.UpdateAudit(ctx, audit)

	reimbursementInfo := s.buildReimbursementInfo(reimbursement)

	audit.WorkflowStatus = WorkflowStatusRAGAudit
	audit.UpdatedAt = time.Now()
	s.repo.UpdateAudit(ctx, audit)

	ragResult, err := s.executeRAGAnalysis(ctx, reimbursementInfo)
	if err != nil {
		s.logger.WithContext(ctx).Error("RAG分析失败", logger.NewField("error", err))
		audit.Status = AuditStatusCompleted
		audit.WorkflowStatus = WorkflowStatusRAGFailed
		audit.Reason = fmt.Sprintf("RAG分析失败: %s", err.Error())
		audit.FinalPass = false
		audit.RAGPass = false
		audit.RiskScore = 1.0
		audit.RiskLevel = "高风险"
		completedTime := time.Now()
		audit.CompletedAt = &completedTime
		audit.Duration = completedTime.Sub(startTime).Milliseconds()
		audit.UpdatedAt = completedTime
		s.repo.UpdateAudit(ctx, audit)

		failFlowLog := &AuditFlowLog{
			ID:              uuid.New().String(),
			ReimbursementID: reimbursementID,
			AuditID:         audit.ID,
			FlowStatus:      FlowStatusIntelligentFail,
			FlowType:        FlowTypeIntelligent,
			Action:          FlowActionRejectAudit,
			Reason:          &audit.Reason,
			CreatedAt:       completedTime,
		}
		s.repo.CreateFlowLog(ctx, failFlowLog)

		return audit, nil
	}

	audit.RAGResults = ragResult
	audit.RAGPass = ragResult != nil && ragResult.Confidence > 0.6

	audit.FinalPass = audit.RulePass && audit.RAGPass
	audit.RiskScore = s.calculateRiskScore(audit)
	audit.RiskLevel = s.determineRiskLevel(audit.RiskScore)
	audit.Suggestions = s.generateSuggestions(audit)
	audit.Reason = s.generateAuditReason(audit)

	completedTime := time.Now()
	audit.CompletedAt = &completedTime
	audit.Duration = completedTime.Sub(startTime).Milliseconds()
	audit.UpdatedAt = completedTime

	if audit.FinalPass {
		audit.WorkflowStatus = WorkflowStatusRAGPassed
		audit.Status = AuditStatusManual
		audit.Reason = "智能审核通过，等待人工审核"

		reimbursement.Status = "auditing"
		if err := s.reimbursementRepo.UpdateReimbursement(ctx, reimbursement); err != nil {
			s.logger.WithContext(ctx).Error("更新报销单状态失败", logger.NewField("error", err))
		}

		passFlowLog := &AuditFlowLog{
			ID:              uuid.New().String(),
			ReimbursementID: reimbursementID,
			AuditID:         audit.ID,
			FlowStatus:      FlowStatusIntelligentPass,
			FlowType:        FlowTypeIntelligent,
			Action:          FlowActionPassAudit,
			CreatedAt:       completedTime,
		}
		if err := s.repo.CreateFlowLog(ctx, passFlowLog); err != nil {
			s.logger.WithContext(ctx).Error("创建流程日志失败", logger.NewField("error", err))
		}

		manualFlowLog := &AuditFlowLog{
			ID:              uuid.New().String(),
			ReimbursementID: reimbursementID,
			AuditID:         audit.ID,
			FlowStatus:      FlowStatusManualStart,
			FlowType:        FlowTypeManual,
			Action:          FlowActionStartAudit,
			CreatedAt:       completedTime,
		}
		if err := s.repo.CreateFlowLog(ctx, manualFlowLog); err != nil {
			s.logger.WithContext(ctx).Error("创建流程日志失败", logger.NewField("error", err))
		}
	} else {
		audit.WorkflowStatus = WorkflowStatusRAGFailed
		audit.Status = AuditStatusCompleted
		audit.Reason = "智能审核未通过"

		reimbursement.Status = "rejected"
		if err := s.reimbursementRepo.UpdateReimbursement(ctx, reimbursement); err != nil {
			s.logger.WithContext(ctx).Error("更新报销单状态失败", logger.NewField("error", err))
		}

		failFlowLog := &AuditFlowLog{
			ID:              uuid.New().String(),
			ReimbursementID: reimbursementID,
			AuditID:         audit.ID,
			FlowStatus:      FlowStatusIntelligentFail,
			FlowType:        FlowTypeIntelligent,
			Action:          FlowActionRejectAudit,
			Reason:          &audit.Reason,
			CreatedAt:       completedTime,
		}
		if err := s.repo.CreateFlowLog(ctx, failFlowLog); err != nil {
			s.logger.WithContext(ctx).Error("创建流程日志失败", logger.NewField("error", err))
		}
	}

	if err := s.repo.UpdateAudit(ctx, audit); err != nil {
		s.logger.WithContext(ctx).Error("更新审核记录失败", logger.NewField("error", err))
		return nil, fmt.Errorf("更新审核记录失败: %w", err)
	}

	s.logger.WithContext(ctx).Info("智能审核完成",
		logger.NewField("audit_id", audit.ID),
		logger.NewField("final_pass", audit.FinalPass),
		logger.NewField("workflow_status", audit.WorkflowStatus),
		logger.NewField("risk_level", audit.RiskLevel),
		logger.NewField("duration", audit.Duration))

	return audit, nil
}

// GetAuditStatus 获取审核状态
func (s *Service) GetAuditStatus(ctx context.Context, auditID string) (*AuditResult, error) {
	audit, err := s.repo.GetAuditByID(ctx, auditID)
	if err != nil {
		s.logger.WithContext(ctx).Error("获取审核记录失败", logger.NewField("error", err))
		return nil, fmt.Errorf("获取审核记录失败: %w", err)
	}

	return audit, nil
}

// GetAuditByReimbursementID 根据报销单ID获取审核结果
func (s *Service) GetAuditByReimbursementID(ctx context.Context, reimbursementID string) (*AuditResult, error) {
	audit, err := s.repo.GetAuditByReimbursementID(ctx, reimbursementID)
	if err != nil {
		s.logger.WithContext(ctx).Error("获取审核记录失败", logger.NewField("error", err))
		return nil, fmt.Errorf("获取审核记录失败: %w", err)
	}

	return audit, nil
}

// executeRuleValidation 执行规则校验
func (s *Service) executeRuleValidation(ctx context.Context, reimbursement *reimbursement.Reimbursement) ([]*RuleValidationResult, error) {
	s.logger.WithContext(ctx).Info("开始规则校验")

	allResults := make([]*RuleValidationResult, 0)

	for _, invoice := range reimbursement.Invoices {
		s.logger.WithContext(ctx).Debug("开始校验发票",
			logger.NewField("发票ID", invoice.ID),
			logger.NewField("发票号码", invoice.Number))

		validationData := map[string]interface{}{
			"invoice":                             invoice,
			"reimbursement":                       reimbursement,
			"apply_date":                          reimbursement.ApplyDate,
			"is_invoice_date_older_than_6_months": invoice.Date != nil && invoice.Date.Before(reimbursement.ApplyDate.AddDate(0, -6, 0)),
			"price":                               invoice.Price,
		}

		results, err := s.ruleEngineService.ExecuteAllRules(ctx, validationData)
		if err != nil {
			s.logger.WithContext(ctx).Error("规则校验失败",
				logger.NewField("发票ID", invoice.ID),
				logger.NewField("error", err))
			return nil, err
		}

		for _, result := range results {
			if !result.Passed && !strings.Contains(result.Message, "未命中") {
				allResults = append(allResults, &RuleValidationResult{
					RuleID:        result.RuleID,
					RuleCode:      result.RuleID,
					RuleName:      result.RuleName,
					RuleType:      result.RuleType,
					Passed:        result.Passed,
					Message:       result.Message,
					Details:       map[string]interface{}{"details": result.Details, "invoice_id": invoice.ID},
					ExecutionTime: result.ExecutionTime,
				})
			}
		}
	}

	s.logger.WithContext(ctx).Info("规则校验完成", logger.NewField("result_count", len(allResults)))

	return allResults, nil
}

// executeRAGAnalysis 执行RAG分析
func (s *Service) executeRAGAnalysis(ctx context.Context, reimbursementInfo map[string]interface{}) (*RAGAnalysisResult, error) {
	s.logger.WithContext(ctx).Info("开始RAG分析")

	result, err := s.ragService.AuditReimbursement(ctx, reimbursementInfo, 5)
	if err != nil {
		s.logger.WithContext(ctx).Error("RAG分析失败", logger.NewField("error", err))
		return nil, err
	}

	ragResult := &RAGAnalysisResult{
		Query:         result.Query,
		Content:       result.AnalysisResult.Conclusion,
		Confidence:    result.AnalysisResult.Confidence,
		Analysis:      result.AnalysisResult.Reasoning,
		ExecutionTime: result.ExecutionTime,
		Chunks:        result.Chunks,
	}

	for _, doc := range result.Documents {
		for _, chunk := range doc.Chunks {
			ragResult.References = append(ragResult.References, &VectorReference{
				ChunkID:    chunk.ID,
				Content:    chunk.Content,
				Similarity: 0.0,
				Category:   doc.Metadata.Category,
				DocumentID: doc.ID,
			})
		}
	}

	s.logger.WithContext(ctx).Info("RAG分析完成", logger.NewField("confidence", ragResult.Confidence))

	return ragResult, nil
}

// buildReimbursementInfo 构建报销单信息（包含发票详细信息）
func (s *Service) buildReimbursementInfo(reimbursement *reimbursement.Reimbursement) map[string]interface{} {
	info := map[string]interface{}{
		"id":            reimbursement.ID,
		"user_id":       reimbursement.UserID,
		"user_name":     reimbursement.UserName,
		"department":    reimbursement.Department,
		"type":          reimbursement.Type,
		"category":      reimbursement.Type,
		"total_amount":  reimbursement.TotalAmount,
		"reason":        reimbursement.TravelReason,
		"description":   reimbursement.Description,
		"apply_date":    reimbursement.ApplyDate,
		"expense_date":  reimbursement.ExpenseDate,
		"invoice_count": len(reimbursement.Invoices),
	}

	if reimbursement.StartDate != nil {
		info["start_date"] = reimbursement.StartDate
	}
	if reimbursement.EndDate != nil {
		info["end_date"] = reimbursement.EndDate
	}
	info["destination"] = reimbursement.Destination
	info["city"] = reimbursement.City
	info["province"] = reimbursement.Province
	info["transportation"] = reimbursement.Transportation
	info["project_code"] = reimbursement.ProjectCode
	info["budget_code"] = reimbursement.BudgetCode

	invoices := make([]map[string]interface{}, len(reimbursement.Invoices))
	for i, invoice := range reimbursement.Invoices {
		invoices[i] = map[string]interface{}{
			"id":                  invoice.ID,
			"type":                invoice.Type,
			"code":                invoice.Code,
			"number":              invoice.Number,
			"date":                invoice.Date,
			"amount":              invoice.Amount,
			"tax_amount":          invoice.TaxAmount,
			"total_amount":        invoice.Amount + invoice.TaxAmount,
			"buyer_name":          invoice.BuyerName,
			"buyer_tax_no":        invoice.BuyerTaxNo,
			"seller_name":         invoice.SellerName,
			"seller_tax_no":       invoice.SellerTaxNo,
			"commodity_name":      invoice.CommodityName,
			"specification":       invoice.Specification,
			"unit":                invoice.Unit,
			"quantity":            invoice.Quantity,
			"price":               invoice.Price,
			"category":            invoice.Category,
			"sub_category":        invoice.SubCategory,
			"expense_type":        invoice.ExpenseType,
			"merchant_type":       invoice.MerchantType,
			"merchant_code":       invoice.MerchantCode,
			"location":            invoice.Location,
			"city":                invoice.City,
			"province":            invoice.Province,
			"country":             invoice.Country,
			"purpose":             invoice.Purpose,
			"description":         invoice.Description,
			"project_code":        invoice.ProjectCode,
			"department_code":     invoice.DepartmentCode,
			"cost_center":         invoice.CostCenter,
			"contract_number":     invoice.ContractNumber,
			"approval_level":      invoice.ApprovalLevel,
			"is_reimbursable":     invoice.IsReimbursable,
			"is_personal":         invoice.IsPersonal,
			"is_vat":              invoice.IsVAT,
			"vat_rate":            invoice.VATRate,
			"exchange_rate":       invoice.ExchangeRate,
			"original_amount":     invoice.OriginalAmount,
			"original_currency":   invoice.OriginalCurrency,
			"is_electronic":       invoice.IsElectronic,
			"is_duplicate":        invoice.IsDuplicate,
			"verification_status": invoice.VerificationStatus,
			"items":               invoice.Items,
			"total_items":         invoice.TotalItems,
			"main_commodity":      invoice.MainCommodity,
		}
	}
	info["invoices"] = invoices

	return info
}

func (s *Service) buildRuleValidationData(reimbursement *reimbursement.Reimbursement) map[string]interface{} {
	return s.buildReimbursementInfo(reimbursement)
}

// checkRulePass 检查规则是否通过
func (s *Service) checkRulePass(results []*RuleValidationResult) bool {
	if len(results) == 0 {
		return true
	}

	for _, result := range results {
		if !result.Passed {
			return false
		}
	}

	return true
}

// calculateRiskScore 计算风险分数
func (s *Service) calculateRiskScore(audit *AuditResult) float64 {
	riskScore := 0.0

	if !audit.RulePass {
		riskScore += 0.5
	}

	if !audit.RAGPass {
		riskScore += 0.3
	}

	if audit.RAGResults != nil {
		riskScore += (1.0 - audit.RAGResults.Confidence) * 0.2
	}

	if riskScore > 1.0 {
		riskScore = 1.0
	}

	return riskScore
}

// determineRiskLevel 确定风险等级
func (s *Service) determineRiskLevel(riskScore float64) string {
	if riskScore >= 0.7 {
		return "高风险"
	} else if riskScore >= 0.4 {
		return "中风险"
	} else {
		return "低风险"
	}
}

// generateSuggestions 生成建议
func (s *Service) generateSuggestions(audit *AuditResult) []string {
	var suggestions []string

	if !audit.RulePass {
		suggestions = append(suggestions, "请检查规则校验不通过的项目")
		for _, result := range audit.RuleResults {
			if !result.Passed {
				suggestions = append(suggestions, fmt.Sprintf("- %s: %s", result.RuleName, result.Message))
			}
		}
	}

	if !audit.RAGPass && audit.RAGResults != nil {
		suggestions = append(suggestions, "请检查RAG分析结果，建议人工复核")
	}

	if audit.RiskLevel == "高风险" {
		suggestions = append(suggestions, "该报销单风险较高，建议进行详细审核")
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "审核通过，可以继续后续流程")
	}

	return suggestions
}

// generateAuditReason 生成审核原因
func (s *Service) generateAuditReason(audit *AuditResult) string {
	if audit.FinalPass {
		return "审核通过"
	}

	var reasons []string

	if !audit.RulePass {
		reasons = append(reasons, "规则校验未通过")
	}

	if !audit.RAGPass {
		reasons = append(reasons, "RAG分析未通过")
	}

	if len(reasons) == 0 {
		return "审核未通过"
	}

	return "审核未通过: " + reasons[0]
}

// WithdrawAudit 撤回报销单审核
func (s *Service) WithdrawAudit(ctx context.Context, reimbursementID string) error {
	s.logger.WithContext(ctx).Info("撤回报销单审核", logger.NewField("reimbursement_id", reimbursementID))

	reimbursement, err := s.reimbursementRepo.GetReimbursementByID(ctx, reimbursementID)
	if err != nil {
		s.logger.WithContext(ctx).Error("获取报销单失败", logger.NewField("error", err))
		return fmt.Errorf("获取报销单失败: %w", err)
	}

	if reimbursement.Status != "auditing" {
		return fmt.Errorf("报销单状态不允许撤回，当前状态: %s", reimbursement.Status)
	}

	auditID := reimbursement.AuditID

	if auditID != "" {
		audit, err := s.repo.GetAuditByReimbursementID(ctx, reimbursementID)
		if err != nil {
			s.logger.WithContext(ctx).Warn("获取审核记录失败", logger.NewField("error", err))
		} else {
			audit.Status = AuditStatusWithdrawn
			audit.Reason = "用户撤回报销单"
			audit.WorkflowStatus = WorkflowStatusWithdrawn
			audit.UpdatedAt = time.Now()
			s.repo.UpdateAudit(ctx, audit)

			withdrawFlowLog := &AuditFlowLog{
				ID:              uuid.New().String(),
				ReimbursementID: reimbursementID,
				AuditID:         audit.ID,
				FlowStatus:      FlowStatusWithdrawn,
				FlowType:        FlowTypeManual,
				Action:          FlowActionWithdrawAudit,
				Reason:          &audit.Reason,
				CreatedAt:       time.Now(),
			}
			s.repo.CreateFlowLog(ctx, withdrawFlowLog)
		}
	}

	reimbursement.Status = "pending_submission"
	reimbursement.AuditID = ""
	if err := s.reimbursementRepo.UpdateReimbursement(ctx, reimbursement); err != nil {
		s.logger.WithContext(ctx).Error("更新报销单状态失败", logger.NewField("error", err))
		return fmt.Errorf("更新报销单状态失败: %w", err)
	}

	s.logger.WithContext(ctx).Info("撤回报销单审核成功", logger.NewField("reimbursement_id", reimbursementID))
	return nil
}

// RetryAudit 重试审核
func (s *Service) RetryAudit(ctx context.Context, auditID string) (*AuditResult, error) {
	audit, err := s.repo.GetAuditByID(ctx, auditID)
	if err != nil {
		return nil, fmt.Errorf("获取审核记录失败: %w", err)
	}

	if audit.Status != AuditStatusFailed {
		return nil, errors.New("只能重试失败的审核")
	}

	return s.StartAudit(ctx, audit.ReimbursementID)
}

// ManualAudit 人工审核
func (s *Service) ManualAudit(ctx context.Context, auditID string, action string, reason string, operatorID string, operatorName string, ipAddress string) (*AuditResult, error) {
	audit, err := s.repo.GetAuditByID(ctx, auditID)
	if err != nil {
		return nil, fmt.Errorf("获取审核记录失败: %w", err)
	}

	if audit.Status != AuditStatusCompleted {
		return nil, errors.New("只能审核已完成的智能审核")
	}

	if !audit.FinalPass {
		return nil, errors.New("智能审核未通过，无法进行人工审核")
	}

	reimbursement, err := s.reimbursementRepo.GetReimbursementByID(ctx, audit.ReimbursementID)
	if err != nil {
		return nil, fmt.Errorf("获取报销单失败: %w", err)
	}

	now := time.Now()
	var flowStatus FlowStatus
	var flowAction FlowAction

	startFlowLog := &AuditFlowLog{
		ID:              uuid.New().String(),
		ReimbursementID: audit.ReimbursementID,
		AuditID:         audit.ID,
		FlowStatus:      FlowStatusManualStart,
		FlowType:        FlowTypeManual,
		OperatorID:      &operatorID,
		OperatorName:    &operatorName,
		Action:          FlowActionStartAudit,
		IPAddress:       &ipAddress,
		CreatedAt:       now,
	}

	if err := s.repo.CreateFlowLog(ctx, startFlowLog); err != nil {
		s.logger.WithContext(ctx).Error("创建流程日志失败", logger.NewField("error", err))
	}

	if action == "pass" {
		flowStatus = FlowStatusManualPass
		flowAction = FlowActionPassAudit
		audit.Status = AuditStatusApproved
		audit.WorkflowStatus = WorkflowStatusManualPassed
		audit.FinalPass = true
		audit.Reason = "人工审核通过"
		if reason != "" {
			audit.Reason += ": " + reason
		}

		reimbursement.Status = "approved"
		if err := s.reimbursementRepo.UpdateReimbursement(ctx, reimbursement); err != nil {
			s.logger.WithContext(ctx).Error("更新报销单状态失败", logger.NewField("error", err))
		}
	} else if action == "reject" {
		flowStatus = FlowStatusManualReject
		flowAction = FlowActionRejectAudit
		audit.Status = AuditStatusRejected
		audit.WorkflowStatus = WorkflowStatusManualRejected
		audit.FinalPass = false
		audit.Reason = "人工审核驳回"
		if reason != "" {
			audit.Reason += ": " + reason
		}

		reimbursement.Status = "rejected"
		if err := s.reimbursementRepo.UpdateReimbursement(ctx, reimbursement); err != nil {
			s.logger.WithContext(ctx).Error("更新报销单状态失败", logger.NewField("error", err))
		}
	} else {
		return nil, errors.New("无效的审核动作")
	}

	audit.UpdatedAt = now

	if err := s.repo.UpdateAudit(ctx, audit); err != nil {
		s.logger.WithContext(ctx).Error("更新审核记录失败", logger.NewField("error", err))
		return nil, fmt.Errorf("更新审核记录失败: %w", err)
	}

	flowLog := &AuditFlowLog{
		ID:              uuid.New().String(),
		ReimbursementID: audit.ReimbursementID,
		AuditID:         audit.ID,
		FlowStatus:      flowStatus,
		FlowType:        FlowTypeManual,
		OperatorID:      &operatorID,
		OperatorName:    &operatorName,
		Action:          flowAction,
		Reason:          &reason,
		IPAddress:       &ipAddress,
		CreatedAt:       now,
	}

	if err := s.repo.CreateFlowLog(ctx, flowLog); err != nil {
		s.logger.WithContext(ctx).Error("创建流程日志失败", logger.NewField("error", err))
		return nil, fmt.Errorf("创建流程日志失败: %w", err)
	}

	s.logger.WithContext(ctx).Info("人工审核完成",
		logger.NewField("audit_id", audit.ID),
		logger.NewField("action", action),
		logger.NewField("operator_id", operatorID))

	return audit, nil
}

// GetFlowLogsByReimbursementID 根据报销单ID获取流程日志
func (s *Service) GetFlowLogsByReimbursementID(ctx context.Context, reimbursementID string) ([]*AuditFlowLog, error) {
	return s.repo.GetFlowLogsByReimbursementID(ctx, reimbursementID)
}

// GetFlowLogsByAuditID 根据审核ID获取流程日志
func (s *Service) GetFlowLogsByAuditID(ctx context.Context, auditID string) ([]*AuditFlowLog, error) {
	return s.repo.GetFlowLogsByAuditID(ctx, auditID)
}
