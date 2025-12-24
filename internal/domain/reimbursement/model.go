// model.go 报销单领域模型
// 功能点：
// 1. 定义报销单数据模型
// 2. 定义发票数据模型
// 3. 定义审核结果数据模型
// 4. 定义审核状态数据模型
// 5. 定义校验结果数据模型
// 6. 提供模型转换和验证方法

package reimbursement

import (
	"time"

	"reimbursement-audit/internal/domain/ocr"
)

// Reimbursement 报销单模型
type Reimbursement struct {
	ID               string         `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`                              // 报销单ID
	UserID           string         `json:"user_id" gorm:"type:varchar(36);not null;column:user_id"`                      // 用户ID
	UserName         string         `json:"user_name" gorm:"type:varchar(100);not null;column:user_name"`                 // 用户姓名
	Department       string         `json:"department" gorm:"type:varchar(100);column:department"`                        // 所属部门
	ApplicantLevel   string         `json:"applicant_level" gorm:"type:varchar(20);column:applicant_level"`               // 申请人级别(高管/经理/员工)
	Type             string         `json:"type" gorm:"type:varchar(50);column:type"`                                     // 报销类型(交通/住宿/餐饮等)
	Title            string         `json:"title" gorm:"type:varchar(200);not null;column:title"`                         // 报销标题
	Description      string         `json:"description" gorm:"type:text;column:description"`                              // 报销描述
	TotalAmount      float64        `json:"total_amount" gorm:"type:decimal(10,2);not null;column:total_amount"`          // 总金额
	Currency         string         `json:"currency" gorm:"type:varchar(10);default:'CNY';column:currency"`               // 币种
	ApplyDate        time.Time      `json:"apply_date" gorm:"type:date;not null;column:apply_date"`                       // 申请日期
	ExpenseDate      time.Time      `json:"expense_date" gorm:"type:date;column:expense_date"`                            // 费用发生日期
	StartDate        time.Time      `json:"start_date" gorm:"type:date;column:start_date"`                                // 出差开始日期
	EndDate          time.Time      `json:"end_date" gorm:"type:date;column:end_date"`                                    // 出差结束日期
	Destination      string         `json:"destination" gorm:"type:varchar(100);column:destination"`                      // 出差目的地
	City             string         `json:"city" gorm:"type:varchar(50);column:city"`                                     // 出差城市
	Province         string         `json:"province" gorm:"type:varchar(50);column:province"`                             // 出差省份
	TravelReason     string         `json:"travel_reason" gorm:"type:varchar(200);column:travel_reason"`                  // 出差事由
	Transportation   string         `json:"transportation" gorm:"type:varchar(50);column:transportation"`                 // 交通工具
	ProjectCode      string         `json:"project_code" gorm:"type:varchar(50);column:project_code"`                     // 项目编码
	BudgetCode       string         `json:"budget_code" gorm:"type:varchar(50);column:budget_code"`                       // 预算科目
	ApprovalRequired bool           `json:"approval_required" gorm:"type:boolean;default:false;column:approval_required"` // 是否需要审批
	ApprovedBy       string         `json:"approved_by" gorm:"type:varchar(36);column:approved_by"`                       // 审批人ID
	ApprovedAt       time.Time      `json:"approved_at" gorm:"type:datetime;column:approved_at"`                          // 审批时间
	Invoices         []*ocr.Invoice `json:"invoices" gorm:"foreignKey:ReimbursementID;constraint:OnDelete:CASCADE"`       // 发票列表
	Status           string         `json:"status" gorm:"type:varchar(20);not null;default:'待提交';column:status"`          // 状态(待提交/待审核/审核中/已完成/已驳回)
	CreatedAt        time.Time      `json:"created_at" gorm:"autoCreateTime"`                                             // 创建时间
	UpdatedAt        time.Time      `json:"updated_at" gorm:"autoUpdateTime"`                                             // 更新时间
	// AuditResults []*AuditResult `json:"audit_results" gorm:"foreignKey:ReimbursementID;constraint:OnDelete:CASCADE"` // 审核结果列表
}

// // AuditResult 审核结果模型
// type AuditResult struct {
// 	ID              string                  `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`                                                      // 审核结果ID
// 	ReimbursementID string                  `json:"reimbursement_id" gorm:"type:varchar(36);not null;index:idx_reimbursement_id;column:reimbursement_id"` // 报销单ID
// 	Status          string                  `json:"status" gorm:"type:varchar(20);not null;column:status"`                                                // 审核状态(待审核/审核中/通过/驳回/需人工复核)
// 	RiskLevel       string                  `json:"risk_level" gorm:"type:varchar(10);column:risk_level"`                                                 // 风险等级(低/中/高)
// 	Score           int                     `json:"score" gorm:"type:int;column:score"`                                                                   // 审核评分
// 	RuleResults     []*RuleValidationResult `json:"rule_results" gorm:"serializer:json;column:rule_results"`                                              // 规则校验结果
// 	RAGResults      *RAGAnalysisResult      `json:"rag_results" gorm:"serializer:json;column:rag_results"`                                                // RAG分析结果
// 	Conclusion      string                  `json:"conclusion" gorm:"type:text;column:conclusion"`                                                        // 审核结论
// 	Reason          string                  `json:"reason" gorm:"type:text;column:reason"`                                                                // 审核理由
// 	Suggestions     []string                `json:"suggestions" gorm:"serializer:json;column:suggestions"`                                                // 审核建议
// 	Issues          []*AuditIssue           `json:"issues" gorm:"serializer:json;column:issues"`                                                          // 审核问题列表
// 	ProcessedBy     string                  `json:"processed_by" gorm:"type:varchar(50);column:processed_by"`                                             // 处理人(系统/人工)
// 	ProcessedAt     time.Time               `json:"processed_at" gorm:"type:datetime;column:processed_at"`                                                // 处理时间
// 	CreatedAt       time.Time               `json:"created_at" gorm:"type:datetime;not null;column:created_at"`                                           // 创建时间
// 	UpdatedAt       time.Time               `json:"updated_at" gorm:"type:datetime;not null;column:updated_at"`                                           // 更新时间
// }

// // AuditStatus 审核状态模型
// type AuditStatus struct {
// 	ID              string    `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`                                                      // 状态ID
// 	ReimbursementID string    `json:"reimbursement_id" gorm:"type:varchar(36);not null;index:idx_reimbursement_id;column:reimbursement_id"` // 报销单ID
// 	Status          string    `json:"status" gorm:"type:varchar(20);not null;column:status"`                                                // 当前状态
// 	Progress        int       `json:"progress" gorm:"type:int;default:0;column:progress"`                                                   // 进度百分比
// 	CurrentStep     string    `json:"current_step" gorm:"type:varchar(100);column:current_step"`                                            // 当前步骤
// 	EstimatedTime   int       `json:"estimated_time" gorm:"type:int;column:estimated_time"`                                                 // 预计完成时间(秒)
// 	StartTime       time.Time `json:"start_time" gorm:"type:datetime;column:start_time"`                                                    // 开始时间
// 	LastUpdateTime  time.Time `json:"last_update_time" gorm:"type:datetime;not null;column:last_update_time"`                               // 最后更新时间
// }

// // ValidationResult 校验结果模型
// type ValidationResult struct {
// 	ID       string   `json:"id" gorm:"primaryKey;type:varchar(36);column:id"` // 校验结果ID
// 	Valid    bool     `json:"valid" gorm:"type:boolean;not null;column:valid"` // 是否通过校验
// 	Errors   []string `json:"errors" gorm:"serializer:json;column:errors"`     // 错误信息
// 	Warnings []string `json:"warnings" gorm:"serializer:json;column:warnings"` // 警告信息
// }

// // RuleValidationResult 规则校验结果模型
// type RuleValidationResult struct {
// 	ID       string `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`              // 规则校验结果ID
// 	RuleID   string `json:"rule_id" gorm:"type:varchar(36);not null;column:rule_id"`      // 规则ID
// 	RuleName string `json:"rule_name" gorm:"type:varchar(100);not null;column:rule_name"` // 规则名称
// 	RuleType string `json:"rule_type" gorm:"type:varchar(50);not null;column:rule_type"`  // 规则类型
// 	Passed   bool   `json:"passed" gorm:"type:boolean;not null;column:passed"`            // 是否通过
// 	Message  string `json:"message" gorm:"type:text;column:message"`                      // 校验消息
// 	Severity string `json:"severity" gorm:"type:varchar(20);column:severity"`             // 严重程度
// 	Details  string `json:"details" gorm:"type:text;column:details"`                      // 详细信息
// 	Priority int    `json:"priority" gorm:"type:int;column:priority"`                     // 优先级
// }

// // RAGAnalysisResult RAG分析结果模型
// type RAGAnalysisResult struct {
// 	ID             string   `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`             // RAG分析结果ID
// 	Query          string   `json:"query" gorm:"type:text;not null;column:query"`                // 查询内容
// 	RetrievedDocs  []string `json:"retrieved_docs" gorm:"serializer:json;column:retrieved_docs"` // 检索到的文档
// 	AnalysisResult string   `json:"analysis_result" gorm:"type:text;column:analysis_result"`     // 分析结果
// 	Confidence     float64  `json:"confidence" gorm:"type:decimal(3,2);column:confidence"`       // 置信度
// }

// // AuditIssue 审核问题模型
// type AuditIssue struct {
// 	ID          string `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`           // 问题ID
// 	Type        string `json:"type" gorm:"type:varchar(50);not null;column:type"`         // 问题类型
// 	Title       string `json:"title" gorm:"type:varchar(200);not null;column:title"`      // 问题标题
// 	Description string `json:"description" gorm:"type:text;column:description"`           // 问题描述
// 	Severity    string `json:"severity" gorm:"type:varchar(20);not null;column:severity"` // 严重程度
// 	Source      string `json:"source" gorm:"type:varchar(20);not null;column:source"`     // 问题来源(规则/RAG)
// 	RuleID      string `json:"rule_id" gorm:"type:varchar(36);column:rule_id"`            // 关联规则ID(如果来源是规则)
// }
