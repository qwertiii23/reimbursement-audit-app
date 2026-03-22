package response

import (
	"reimbursement-audit/internal/domain/audit"
	"time"
)

// AuditResponse 审核响应
type AuditResponse struct {
	ID              string     `json:"id"`
	ReimbursementID string     `json:"reimbursement_id"`
	Status          string     `json:"status"`
	WorkflowStatus  string     `json:"workflow_status"`
	RulePass        bool       `json:"rule_pass"`
	RAGPass         bool       `json:"rag_pass"`
	FinalPass       bool       `json:"final_pass"`
	RiskLevel       string     `json:"risk_level"`
	RiskScore       float64    `json:"risk_score"`
	Reason          string     `json:"reason"`
	Suggestions     []string   `json:"suggestions"`
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	Duration        int64      `json:"duration"`
}

// AuditStatusResponse 审核状态响应
type AuditStatusResponse struct {
	ID              string     `json:"id"`
	ReimbursementID string     `json:"reimbursement_id"`
	Status          string     `json:"status"`
	WorkflowStatus  string     `json:"workflow_status"`
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	Duration        int64      `json:"duration"`
}

// AuditResultResponse 审核结果响应
type AuditResultResponse struct {
	ID              string                     `json:"id"`
	ReimbursementID string                     `json:"reimbursement_id"`
	Status          string                     `json:"status"`
	WorkflowStatus  string                     `json:"workflow_status"`
	RulePass        bool                       `json:"rule_pass"`
	RAGPass         bool                       `json:"rag_pass"`
	FinalPass       bool                       `json:"final_pass"`
	RuleResults     []*RuleValidationResult    `json:"rule_results"`
	RAGResults      *RAGAnalysisResultResponse `json:"rag_results"`
	RiskLevel       string                     `json:"risk_level"`
	RiskScore       float64                    `json:"risk_score"`
	Reason          string                     `json:"reason"`
	Suggestions     []string                   `json:"suggestions"`
	StartedAt       time.Time                  `json:"started_at"`
	CompletedAt     *time.Time                 `json:"completed_at"`
	Duration        int64                      `json:"duration"`
}

// RuleValidationResult 规则校验结果响应
type RuleValidationResult struct {
	RuleID        string                 `json:"rule_id"`
	RuleCode      string                 `json:"rule_code"`
	RuleName      string                 `json:"rule_name"`
	RuleType      string                 `json:"rule_type"`
	Passed        bool                   `json:"passed"`
	Message       string                 `json:"message"`
	Details       map[string]interface{} `json:"details"`
	ExecutionTime int64                  `json:"execution_time"`
}

// RAGAnalysisResultResponse RAG分析结果响应
type RAGAnalysisResultResponse struct {
	Query         string             `json:"query"`
	Content       string             `json:"content"`
	Confidence    float64            `json:"confidence"`
	References    []*VectorReference `json:"references"`
	Analysis      string             `json:"analysis"`
	ExecutionTime int64              `json:"execution_time"`
}

// VectorReference 向量检索引用响应
type VectorReference struct {
	ChunkID    string  `json:"chunk_id"`
	Content    string  `json:"content"`
	Similarity float64 `json:"similarity"`
	Category   string  `json:"category"`
	DocumentID string  `json:"document_id"`
}

// NewAuditResponse 创建审核响应
func NewAuditResponse(auditResult *audit.AuditResult) *AuditResponse {
	return &AuditResponse{
		ID:              auditResult.ID,
		ReimbursementID: auditResult.ReimbursementID,
		Status:          string(auditResult.Status),
		WorkflowStatus:  string(auditResult.WorkflowStatus),
		RulePass:        auditResult.RulePass,
		RAGPass:         auditResult.RAGPass,
		FinalPass:       auditResult.FinalPass,
		RiskLevel:       auditResult.RiskLevel,
		RiskScore:       auditResult.RiskScore,
		Reason:          auditResult.Reason,
		Suggestions:     auditResult.Suggestions,
		StartedAt:       auditResult.StartedAt,
		CompletedAt:     auditResult.CompletedAt,
		Duration:        auditResult.Duration,
	}
}

// NewAuditStatusResponse 创建审核状态响应
func NewAuditStatusResponse(auditResult *audit.AuditResult) *AuditStatusResponse {
	return &AuditStatusResponse{
		ID:              auditResult.ID,
		ReimbursementID: auditResult.ReimbursementID,
		Status:          string(auditResult.Status),
		WorkflowStatus:  string(auditResult.WorkflowStatus),
		StartedAt:       auditResult.StartedAt,
		CompletedAt:     auditResult.CompletedAt,
		Duration:        auditResult.Duration,
	}
}

// NewAuditResultResponse 创建审核结果响应
func NewAuditResultResponse(auditResult *audit.AuditResult) *AuditResultResponse {
	response := &AuditResultResponse{
		ID:              auditResult.ID,
		ReimbursementID: auditResult.ReimbursementID,
		Status:          string(auditResult.Status),
		WorkflowStatus:  string(auditResult.WorkflowStatus),
		RulePass:        auditResult.RulePass,
		RAGPass:         auditResult.RAGPass,
		FinalPass:       auditResult.FinalPass,
		RiskLevel:       auditResult.RiskLevel,
		RiskScore:       auditResult.RiskScore,
		Reason:          auditResult.Reason,
		Suggestions:     auditResult.Suggestions,
		StartedAt:       auditResult.StartedAt,
		CompletedAt:     auditResult.CompletedAt,
		Duration:        auditResult.Duration,
	}

	if auditResult.RuleResults != nil {
		response.RuleResults = make([]*RuleValidationResult, len(auditResult.RuleResults))
		for i, result := range auditResult.RuleResults {
			response.RuleResults[i] = &RuleValidationResult{
				RuleID:        result.RuleID,
				RuleCode:      result.RuleCode,
				RuleName:      result.RuleName,
				RuleType:      result.RuleType,
				Passed:        result.Passed,
				Message:       result.Message,
				Details:       result.Details,
				ExecutionTime: result.ExecutionTime,
			}
		}
	}

	if auditResult.RAGResults != nil {
		response.RAGResults = &RAGAnalysisResultResponse{
			Query:         auditResult.RAGResults.Query,
			Content:       auditResult.RAGResults.Content,
			Confidence:    auditResult.RAGResults.Confidence,
			Analysis:      auditResult.RAGResults.Analysis,
			ExecutionTime: auditResult.RAGResults.ExecutionTime,
		}

		if auditResult.RAGResults.References != nil {
			response.RAGResults.References = make([]*VectorReference, len(auditResult.RAGResults.References))
			for i, ref := range auditResult.RAGResults.References {
				response.RAGResults.References[i] = &VectorReference{
					ChunkID:    ref.ChunkID,
					Content:    ref.Content,
					Similarity: ref.Similarity,
					Category:   ref.Category,
					DocumentID: ref.DocumentID,
				}
			}
		}
	}

	return response
}

// FlowLogResponse 流程日志响应
type FlowLogResponse struct {
	ID              string    `json:"id"`
	ReimbursementID string    `json:"reimbursement_id"`
	AuditID         string    `json:"audit_id"`
	FlowStatus      string    `json:"flow_status"`
	FlowType        string    `json:"flow_type"`
	OperatorID      *string   `json:"operator_id"`
	OperatorName    *string   `json:"operator_name"`
	Action          string    `json:"action"`
	Reason          *string   `json:"reason"`
	Result          *string   `json:"result"`
	IPAddress       *string   `json:"ip_address"`
	CreatedAt       time.Time `json:"created_at"`
}

// NewFlowLogResponse 创建流程日志响应
func NewFlowLogResponse(flowLog *audit.AuditFlowLog) *FlowLogResponse {
	return &FlowLogResponse{
		ID:              flowLog.ID,
		ReimbursementID: flowLog.ReimbursementID,
		AuditID:         flowLog.AuditID,
		FlowStatus:      string(flowLog.FlowStatus),
		FlowType:        string(flowLog.FlowType),
		OperatorID:      flowLog.OperatorID,
		OperatorName:    flowLog.OperatorName,
		Action:          string(flowLog.Action),
		Reason:          flowLog.Reason,
		Result:          flowLog.Result,
		IPAddress:       flowLog.IPAddress,
		CreatedAt:       flowLog.CreatedAt,
	}
}
