// model.go 报销单领域模型
// 功能点：
// 1. 定义报销单数据模型
// 2. 定义发票数据模型
// 3. 定义审核结果数据模型
// 4. 定义审核状态数据模型
// 5. 定义校验结果数据模型
// 6. 提供模型转换和验证方法

package reimbursement

import "time"

// Reimbursement 报销单模型
type Reimbursement struct {
	ID           string         `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`                             // 报销单ID
	UserID       string         `json:"user_id" gorm:"type:varchar(36);not null;column:user_id"`                     // 用户ID
	UserName     string         `json:"user_name" gorm:"type:varchar(100);not null;column:user_name"`                // 用户姓名
	Department   string         `json:"department" gorm:"type:tvarchar(100);column:department"`                       // 所属部门
	Type         string         `json:"type" gorm:"type:varchar(50);column:type"`                                    // 报销类型(交通/住宿/餐饮等)
	Title        string         `json:"title" gorm:"type:varchar(200);not null;column:title"`                        // 报销标题
	Description  string         `json:"description" gorm:"type:text;column:description"`                             // 报销描述
	TotalAmount  float64        `json:"total_amount" gorm:"type:decimal(10,2);not null;column:total_amount"`         // 总金额
	Currency     string         `json:"currency" gorm:"type:varchar(10);default:'CNY';column:currency"`              // 币种
	ApplyDate    time.Time      `json:"apply_date" gorm:"type:date;not null;column:apply_date"`                      // 申请日期
	ExpenseDate  time.Time      `json:"expense_date" gorm:"type:date;column:expense_date"`                           // 费用发生日期
	Invoices     []*Invoice     `json:"invoices" gorm:"foreignKey:ReimbursementID;constraint:OnDelete:CASCADE"`      // 发票列表
	Status       string         `json:"status" gorm:"type:varchar(20);not null;default:'待提交';column:status"`         // 状态(待提交/待审核/审核中/已完成/已驳回)
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`                  // 创建时间
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`                  // 更新时间
	AuditResults []*AuditResult `json:"audit_results" gorm:"foreignKey:ReimbursementID;constraint:OnDelete:CASCADE"` // 审核结果列表
}

// Invoice 发票模型
type Invoice struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`                                                      // 发票ID
	ReimbursementID string    `json:"reimbursement_id" gorm:"type:varchar(36);not null;index:idx_reimbursement_id;column:reimbursement_id"` // 报销单ID
	Type            string    `json:"type" gorm:"type:varchar(50);column:type"`                                                             // 发票类型(增值税发票/定额发票等)
	Code            string    `json:"code" gorm:"type:varchar(50);column:code"`                                                             // 发票代码
	Number          string    `json:"number" gorm:"type:varchar(50);column:number"`                                                         // 发票号码
	Date            time.Time `json:"date" gorm:"type:date;column:date"`                                                                    // 开票日期
	Amount          float64   `json:"amount" gorm:"type:decimal(10,2);not null;column:amount"`                                              // 发票金额
	TaxAmount       float64   `json:"tax_amount" gorm:"type:decimal(10,2);column:tax_amount"`                                               // 税额
	Payer           string    `json:"payer" gorm:"type:varchar(100);column:payer"`                                                          // 付款方
	Payee           string    `json:"payee" gorm:"type:varchar(100);column:payee"`                                                          // 收款方
	BuyerName       string    `json:"buyer_name" gorm:"type:varchar(100);column:buyer_name"`                                                // 购买方名称
	BuyerTaxNo      string    `json:"buyer_tax_no" gorm:"type:varchar(50);column:buyer_tax_no"`                                             // 购买方税号
	SellerName      string    `json:"seller_name" gorm:"type:varchar(100);column:seller_name"`                                              // 销售方名称
	SellerTaxNo     string    `json:"seller_tax_no" gorm:"type:varchar(50);column:seller_tax_no"`                                           // 销售方税号
	CommodityName   string    `json:"commodity_name" gorm:"type:varchar(200);column:commodity_name"`                                        // 商品名称
	Specification   string    `json:"specification" gorm:"type:varchar(100);column:specification"`                                          // 规格型号
	Unit            string    `json:"unit" gorm:"type:varchar(20);column:unit"`                                                             // 单位
	Quantity        float64   `json:"quantity" gorm:"type:decimal(10,2);column:quantity"`                                                   // 数量
	Price           float64   `json:"price" gorm:"type:decimal(10,2);column:price"`                                                         // 单价
	ImagePath       string    `json:"image_path" gorm:"type:varchar(500);column:image_path"`                                                // 发票图片路径
	OCRResult       string    `json:"ocr_result" gorm:"type:text;column:ocr_result"`                                                        // OCR识别结果
	Status          string    `json:"status" gorm:"type:varchar(20);not null;default:'待识别';column:status"`                                  // 状态(待识别/已识别/识别失败)
	CreatedAt       time.Time `json:"created_at" gorm:"type:datetime;not null;column:created_at"`                                           // 创建时间
	UpdatedAt       time.Time `json:"updated_at" gorm:"type:datetime;not null;column:updated_at"`                                           // 更新时间
}

// AuditResult 审核结果模型
type AuditResult struct {
	ID              string                  `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`                                                      // 审核结果ID
	ReimbursementID string                  `json:"reimbursement_id" gorm:"type:varchar(36);not null;index:idx_reimbursement_id;column:reimbursement_id"` // 报销单ID
	Status          string                  `json:"status" gorm:"type:varchar(20);not null;column:status"`                                                // 审核状态(待审核/审核中/通过/驳回/需人工复核)
	RiskLevel       string                  `json:"risk_level" gorm:"type:varchar(10);column:risk_level"`                                                 // 风险等级(低/中/高)
	Score           int                     `json:"score" gorm:"type:int;column:score"`                                                                   // 审核评分
	RuleResults     []*RuleValidationResult `json:"rule_results" gorm:"serializer:json;column:rule_results"`                                              // 规则校验结果
	RAGResults      *RAGAnalysisResult      `json:"rag_results" gorm:"serializer:json;column:rag_results"`                                                // RAG分析结果
	Conclusion      string                  `json:"conclusion" gorm:"type:text;column:conclusion"`                                                        // 审核结论
	Reason          string                  `json:"reason" gorm:"type:text;column:reason"`                                                                // 审核理由
	Suggestions     []string                `json:"suggestions" gorm:"serializer:json;column:suggestions"`                                                // 审核建议
	Issues          []*AuditIssue           `json:"issues" gorm:"serializer:json;column:issues"`                                                          // 审核问题列表
	ProcessedBy     string                  `json:"processed_by" gorm:"type:varchar(50);column:processed_by"`                                             // 处理人(系统/人工)
	ProcessedAt     time.Time               `json:"processed_at" gorm:"type:datetime;column:processed_at"`                                                // 处理时间
	CreatedAt       time.Time               `json:"created_at" gorm:"type:datetime;not null;column:created_at"`                                           // 创建时间
	UpdatedAt       time.Time               `json:"updated_at" gorm:"type:datetime;not null;column:updated_at"`                                           // 更新时间
}

// AuditStatus 审核状态模型
type AuditStatus struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`                                                      // 状态ID
	ReimbursementID string    `json:"reimbursement_id" gorm:"type:varchar(36);not null;index:idx_reimbursement_id;column:reimbursement_id"` // 报销单ID
	Status          string    `json:"status" gorm:"type:varchar(20);not null;column:status"`                                                // 当前状态
	Progress        int       `json:"progress" gorm:"type:int;default:0;column:progress"`                                                   // 进度百分比
	CurrentStep     string    `json:"current_step" gorm:"type:varchar(100);column:current_step"`                                            // 当前步骤
	EstimatedTime   int       `json:"estimated_time" gorm:"type:int;column:estimated_time"`                                                 // 预计完成时间(秒)
	StartTime       time.Time `json:"start_time" gorm:"type:datetime;column:start_time"`                                                    // 开始时间
	LastUpdateTime  time.Time `json:"last_update_time" gorm:"type:datetime;not null;column:last_update_time"`                               // 最后更新时间
}

// ValidationResult 校验结果模型
type ValidationResult struct {
	ID       string   `json:"id" gorm:"primaryKey;type:varchar(36);column:id"` // 校验结果ID
	Valid    bool     `json:"valid" gorm:"type:boolean;not null;column:valid"` // 是否通过校验
	Errors   []string `json:"errors" gorm:"serializer:json;column:errors"`     // 错误信息
	Warnings []string `json:"warnings" gorm:"serializer:json;column:warnings"` // 警告信息
}

// RuleValidationResult 规则校验结果模型
type RuleValidationResult struct {
	ID       string `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`              // 规则校验结果ID
	RuleID   string `json:"rule_id" gorm:"type:varchar(36);not null;column:rule_id"`      // 规则ID
	RuleName string `json:"rule_name" gorm:"type:varchar(100);not null;column:rule_name"` // 规则名称
	RuleType string `json:"rule_type" gorm:"type:varchar(50);not null;column:rule_type"`  // 规则类型
	Passed   bool   `json:"passed" gorm:"type:boolean;not null;column:passed"`            // 是否通过
	Message  string `json:"message" gorm:"type:text;column:message"`                      // 校验消息
	Severity string `json:"severity" gorm:"type:varchar(20);column:severity"`             // 严重程度
	Details  string `json:"details" gorm:"type:text;column:details"`                      // 详细信息
	Priority int    `json:"priority" gorm:"type:int;column:priority"`                     // 优先级
}

// RAGAnalysisResult RAG分析结果模型
type RAGAnalysisResult struct {
	ID             string   `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`             // RAG分析结果ID
	Query          string   `json:"query" gorm:"type:text;not null;column:query"`                // 查询内容
	RetrievedDocs  []string `json:"retrieved_docs" gorm:"serializer:json;column:retrieved_docs"` // 检索到的文档
	AnalysisResult string   `json:"analysis_result" gorm:"type:text;column:analysis_result"`     // 分析结果
	Confidence     float64  `json:"confidence" gorm:"type:decimal(3,2);column:confidence"`       // 置信度
}

// AuditIssue 审核问题模型
type AuditIssue struct {
	ID          string `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`           // 问题ID
	Type        string `json:"type" gorm:"type:varchar(50);not null;column:type"`         // 问题类型
	Title       string `json:"title" gorm:"type:varchar(200);not null;column:title"`      // 问题标题
	Description string `json:"description" gorm:"type:text;column:description"`           // 问题描述
	Severity    string `json:"severity" gorm:"type:varchar(20);not null;column:severity"` // 严重程度
	Source      string `json:"source" gorm:"type:varchar(20);not null;column:source"`     // 问题来源(规则/RAG)
	RuleID      string `json:"rule_id" gorm:"type:varchar(36);column:rule_id"`            // 关联规则ID(如果来源是规则)
}

// IsValid 检查报销单是否有效
func (r *Reimbursement) IsValid() bool {
	// TODO: 实现报销单有效性检查逻辑
	return false
}

// IsValid 检查发票是否有效
func (i *Invoice) IsValid() bool {
	// TODO: 实现发票有效性检查逻辑
	return false
}

// IsPassed 检查审核是否通过
func (a *AuditResult) IsPassed() bool {
	// TODO: 实现审核通过检查逻辑
	return false
}

// GetFailedRules 获取失败的规则列表
func (a *AuditResult) GetFailedRules() []*RuleValidationResult {
	// TODO: 实现获取失败规则列表逻辑
	return nil
}

// GetHighRiskIssues 获取高风险问题列表
func (a *AuditResult) GetHighRiskIssues() []*AuditIssue {
	// TODO: 实现获取高风险问题列表逻辑
	return nil
}

