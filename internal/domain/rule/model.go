// model.go 规则模型（规则定义、优先级）
// 功能点：
// 1. 定义规则数据模型
// 2. 定义规则校验结果模型
// 3. 定义规则过滤器模型
// 4. 定义规则类型枚举
// 5. 定义规则优先级枚举
// 6. 提供模型转换和验证方法

package rule

import "time"

// Rule 规则模型
type Rule struct {
	ID          string    `json:"id"`          // 规则ID
	Name        string    `json:"name"`        // 规则名称
	Description string    `json:"description"` // 规则描述
	Type        string    `json:"type"`        // 规则类型(金额/频次/发票/合规等)
	Category    string    `json:"category"`    // 规则分类
	Definition  string    `json:"definition"`  // 规则定义(Grule语法)
	Priority    int       `json:"priority"`    // 优先级(数字越大优先级越高)
	Enabled     bool      `json:"enabled"`     // 是否启用
	CreatedBy   string    `json:"created_by"`  // 创建人
	UpdatedBy   string    `json:"updated_by"`  // 更新人
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`  // 更新时间
	Version     int       `json:"version"`     // 版本号
	Tags        []string  `json:"tags"`        // 标签
	Metadata    map[string]interface{} `json:"metadata"` // 元数据
}

// RuleValidationResult 规则校验结果模型
type RuleValidationResult struct {
	RuleID      string                 `json:"rule_id"`      // 规则ID
	RuleName    string                 `json:"rule_name"`    // 规则名称
	RuleType    string                 `json:"rule_type"`    // 规则类型
	Passed      bool                   `json:"passed"`       // 是否通过
	Message     string                 `json:"message"`      // 校验消息
	Details     string                 `json:"details"`      // 详细信息
	Severity    string                 `json:"severity"`     // 严重程度(低/中/高)
	Priority    int                    `json:"priority"`     // 优先级
	ExecutionTime int64                `json:"execution_time"` // 执行时间(毫秒)
	Data        map[string]interface{} `json:"data"`         // 相关数据
	Timestamp   time.Time              `json:"timestamp"`    // 校验时间
}

// RuleFilter 规则过滤器模型
type RuleFilter struct {
	Type     string   `json:"type"`     // 规则类型
	Category string   `json:"category"` // 规则分类
	Enabled  *bool    `json:"enabled"`  // 是否启用
	Tags     []string `json:"tags"`     // 标签
	Page     int      `json:"page"`     // 页码
	Size     int      `json:"size"`     // 每页大小
}

// RuleStatistics 规则统计模型
type RuleStatistics struct {
	RuleID         string  `json:"rule_id"`         // 规则ID
	ExecutionCount int64   `json:"execution_count"` // 执行次数
	PassCount      int64   `json:"pass_count"`      // 通过次数
	FailCount      int64   `json:"fail_count"`      // 失败次数
	PassRate       float64 `json:"pass_rate"`       // 通过率
	AvgTime        float64 `json:"avg_time"`        // 平均执行时间(毫秒)
	LastExecuted   time.Time `json:"last_executed"` // 最后执行时间
}

// RuleExecutionLog 规则执行日志模型
type RuleExecutionLog struct {
	ID            string                 `json:"id"`            // 日志ID
	RuleID        string                 `json:"rule_id"`       // 规则ID
	ReimbursementID string               `json:"reimbursement_id"` // 报销单ID
	InputData     map[string]interface{} `json:"input_data"`    // 输入数据
	OutputData    map[string]interface{} `json:"output_data"`   // 输出数据
	Passed        bool                   `json:"passed"`         // 是否通过
	Message       string                 `json:"message"`        // 执行消息
	ExecutionTime int64                  `json:"execution_time"` // 执行时间(毫秒)
	Timestamp     time.Time              `json:"timestamp"`     // 执行时间
}

// RuleType 规则类型常量
const (
	RuleTypeAmount       = "amount"       // 金额校验规则
	RuleTypeFrequency    = "frequency"    // 频次校验规则
	RuleTypeInvoice      = "invoice"      // 发票校验规则
	RuleTypeCompliance   = "compliance"   // 合规校验规则
	RuleTypeCustom       = "custom"       // 自定义规则
)

// RuleCategory 规则分类常量
const (
	RuleCategoryTravel      = "travel"      // 差旅报销规则
	RuleCategoryDaily       = "daily"       // 日常费用规则
	RuleCategoryEntertainment = "entertainment" // 招待费规则
	RuleCategoryOffice      = "office"      // 办公用品规则
	RuleCategoryCommunication = "communication" // 通讯费规则
	RuleCategoryOther       = "other"       // 其他费用规则
)

// RuleSeverity 规则严重程度常量
const (
	RuleSeverityLow    = "low"    // 低严重程度
	RuleSeverityMedium = "medium" // 中等严重程度
	RuleSeverityHigh   = "high"   // 高严重程度
)

// IsValid 检查规则是否有效
func (r *Rule) IsValid() bool {
	// TODO: 实现规则有效性检查逻辑
	return false
}

// IsHighPriority 检查规则是否为高优先级
func (r *Rule) IsHighPriority() bool {
	// TODO: 实现高优先级检查逻辑
	return false
}

// GetSeverityLevel 获取严重程度级别
func (r *RuleValidationResult) GetSeverityLevel() int {
	// TODO: 实现获取严重程度级别逻辑
	return 0
}

// IsFailed 检查规则校验是否失败
func (r *RuleValidationResult) IsFailed() bool {
	// TODO: 实现规则校验失败检查逻辑
	return false
}

// UpdateStatistics 更新统计信息
func (s *RuleStatistics) UpdateStatistics(passed bool, executionTime int64) {
	// TODO: 实现更新统计信息逻辑
}

// CalculatePassRate 计算通过率
func (s *RuleStatistics) CalculatePassRate() float64 {
	// TODO: 实现计算通过率逻辑
	return 0
}

// CalculateAvgTime 计算平均执行时间
func (s *RuleStatistics) CalculateAvgTime() float64 {
	// TODO: 实现计算平均执行时间逻辑
	return 0
}