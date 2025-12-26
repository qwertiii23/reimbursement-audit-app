package audit

import (
	"reimbursement-audit/internal/domain/rag"
	"time"
)

// AuditStatus 审核状态
type AuditStatus string

const (
	AuditStatusPending   AuditStatus = "待审核"
	AuditStatusRunning   AuditStatus = "审核中"
	AuditStatusCompleted AuditStatus = "审核完成"
	AuditStatusFailed    AuditStatus = "审核失败"
)

// AuditResult 审核结果
type AuditResult struct {
	ID              string                  `json:"id"`
	ReimbursementID string                  `json:"reimbursement_id"`
	Status          AuditStatus             `json:"status"`
	RulePass        bool                    `json:"rule_pass"`
	RAGPass         bool                    `json:"rag_pass"`
	FinalPass       bool                    `json:"final_pass"`
	RuleResults     []*RuleValidationResult `json:"rule_results"`
	RAGResults      *RAGAnalysisResult      `json:"rag_results"`
	RiskLevel       string                  `json:"risk_level"`
	RiskScore       float64                 `json:"risk_score"`
	Reason          string                  `json:"reason"`
	Suggestions     []string                `json:"suggestions"`
	StartedAt       time.Time               `json:"started_at"`
	CompletedAt     *time.Time              `json:"completed_at"`
	Duration        int64                   `json:"duration"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

// RuleValidationResult 规则校验结果
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

// RAGAnalysisResult RAG分析结果
type RAGAnalysisResult struct {
	Query         string               `json:"query"`
	Content       string               `json:"content"`
	Confidence    float64              `json:"confidence"`
	References    []*VectorReference   `json:"references"`
	Analysis      string               `json:"analysis"`
	ExecutionTime int64                `json:"execution_time"`
	Chunks        []*rag.DocumentChunk `json:"chunks"`
}

// VectorReference 向量检索引用
type VectorReference struct {
	ChunkID    string  `json:"chunk_id"`
	Content    string  `json:"content"`
	Similarity float64 `json:"similarity"`
	Category   string  `json:"category"`
	DocumentID string  `json:"document_id"`
}

// AuditFilter 审核查询过滤器
type AuditFilter struct {
	ReimbursementID string      `json:"reimbursement_id"`
	Status          AuditStatus `json:"status"`
	StartTime       *time.Time  `json:"start_time"`
	EndTime         *time.Time  `json:"end_time"`
	Page            int         `json:"page"`
	Size            int         `json:"size"`
}
