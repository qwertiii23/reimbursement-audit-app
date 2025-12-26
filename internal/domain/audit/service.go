package audit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"reimbursement-audit/internal/domain/rag"
	"reimbursement-audit/internal/domain/reimbursement"
	"reimbursement-audit/internal/domain/rule"
	"reimbursement-audit/internal/pkg/logger"

	"github.com/google/uuid"
)

// Service 审核服务
type Service struct {
	repo              Repository
	reimbursementRepo reimbursement.Repository
	ruleService       *rule.RuleService
	ragService        *rag.RAGService
	logger            logger.Logger
}

// NewService 创建审核服务
func NewService(
	repo Repository,
	reimbursementRepo reimbursement.Repository,
	ruleService *rule.RuleService,
	ragService *rag.RAGService,
	logger logger.Logger,
) *Service {
	return &Service{
		repo:              repo,
		reimbursementRepo: reimbursementRepo,
		ruleService:       ruleService,
		ragService:        ragService,
		logger:            logger,
	}
}

// StartAudit 开始审核
func (s *Service) StartAudit(ctx context.Context, reimbursementID string) (*AuditResult, error) {
	startTime := time.Now()

	s.logger.WithContext(ctx).Info("开始审核", logger.NewField("reimbursement_id", reimbursementID))

	reimbursement, err := s.reimbursementRepo.GetReimbursementByID(ctx, reimbursementID)
	if err != nil {
		s.logger.WithContext(ctx).Error("获取报销单失败", logger.NewField("error", err))
		return nil, fmt.Errorf("获取报销单失败: %w", err)
	}

	audit := &AuditResult{
		ID:              uuid.New().String(),
		ReimbursementID: reimbursementID,
		Status:          AuditStatusRunning,
		StartedAt:       startTime,
		CreatedAt:       startTime,
		UpdatedAt:       startTime,
	}

	if err := s.repo.CreateAudit(ctx, audit); err != nil {
		s.logger.WithContext(ctx).Error("创建审核记录失败", logger.NewField("error", err))
		return nil, fmt.Errorf("创建审核记录失败: %w", err)
	}

	ruleResults, err := s.executeRuleValidation(ctx, reimbursement)
	if err != nil {
		s.logger.WithContext(ctx).Error("规则校验失败", logger.NewField("error", err))
		audit.Status = AuditStatusFailed
		audit.Reason = fmt.Sprintf("规则校验失败: %s", err.Error())
		audit.CompletedAt = &startTime
		audit.Duration = time.Since(startTime).Milliseconds()
		s.repo.UpdateAudit(ctx, audit)
		return nil, err
	}

	audit.RuleResults = ruleResults
	rulePass := s.checkRulePass(ruleResults)
	audit.RulePass = rulePass

	reimbursementInfo := s.buildReimbursementInfo(reimbursement)
	ragResult, err := s.executeRAGAnalysis(ctx, reimbursementInfo)
	if err != nil {
		s.logger.WithContext(ctx).Error("RAG分析失败", logger.NewField("error", err))
		audit.Status = AuditStatusFailed
		audit.Reason = fmt.Sprintf("RAG分析失败: %s", err.Error())
		audit.CompletedAt = &startTime
		audit.Duration = time.Since(startTime).Milliseconds()
		s.repo.UpdateAudit(ctx, audit)
		return nil, err
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
	audit.Status = AuditStatusCompleted
	audit.UpdatedAt = completedTime

	if err := s.repo.UpdateAudit(ctx, audit); err != nil {
		s.logger.WithContext(ctx).Error("更新审核记录失败", logger.NewField("error", err))
		return nil, fmt.Errorf("更新审核记录失败: %w", err)
	}

	s.logger.WithContext(ctx).Info("审核完成",
		logger.NewField("audit_id", audit.ID),
		logger.NewField("final_pass", audit.FinalPass),
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

	data := s.buildRuleValidationData(reimbursement)
	results, err := s.ruleService.ValidateAllRules(ctx, data)
	if err != nil {
		s.logger.WithContext(ctx).Error("规则校验失败", logger.NewField("error", err))
		return nil, err
	}

	convertedResults := make([]*RuleValidationResult, len(results))
	for i, result := range results {
		convertedResults[i] = &RuleValidationResult{
			RuleID:        result.RuleID,
			RuleCode:      result.RuleID,
			RuleName:      result.RuleName,
			RuleType:      result.RuleType,
			Passed:        result.Passed,
			Message:       result.Message,
			Details:       map[string]interface{}{"details": result.Details},
			ExecutionTime: result.ExecutionTime,
		}
	}

	s.logger.WithContext(ctx).Info("规则校验完成", logger.NewField("result_count", len(results)))

	return convertedResults, nil
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

// buildReimbursementInfo 构建报销单信息
func (s *Service) buildReimbursementInfo(reimbursement *reimbursement.Reimbursement) map[string]interface{} {
	return map[string]interface{}{
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
}

// buildRuleValidationData 构建规则校验数据
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
