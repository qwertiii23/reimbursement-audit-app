// invoice_validator.go 发票校验器
// 功能点：
// 1. 定义发票校验结果模型
// 2. 定义发票校验规则接口
// 3. 实现基础刚性规则校验逻辑
// 4. 提供规则优先级执行和错误聚合功能

package rule

import (
	"context"
	"errors"
	"time"

	"reimbursement-audit/internal/domain/ocr"
	"reimbursement-audit/internal/domain/reimbursement"
	"reimbursement-audit/internal/pkg/logger"
)

// InvoiceValidationResult 发票校验结果
type InvoiceValidationResult struct {
	Passed     bool                `json:"passed"`     // 是否通过校验
	InvoiceID  string              `json:"invoice_id"` // 发票ID
	Violations []*InvoiceViolation `json:"violations"` // 违规规则列表
	Summary    string              `json:"summary"`    // 校验结果摘要
	Timestamp  time.Time           `json:"timestamp"`  // 校验时间
}

// InvoiceViolation 发票违规信息
type InvoiceViolation struct {
	RuleID     string `json:"rule_id"`    // 规则ID
	RuleName   string `json:"rule_name"`  // 规则名称
	RuleType   string `json:"rule_type"`  // 规则类型
	Severity   string `json:"severity"`   // 严重程度(高/中/低)
	Message    string `json:"message"`    // 违规描述
	Suggestion string `json:"suggestion"` // 修改建议
	Priority   int    `json:"priority"`   // 规则优先级
}

// InvoiceValidationRequest 发票校验请求
type InvoiceValidationRequest struct {
	Invoice       *ocr.Invoice                 `json:"invoice"`       // 待校验发票
	Reimbursement *reimbursement.Reimbursement `json:"reimbursement"` // 关联报销单
	CompanyNames  []string                     `json:"company_names"` // 允许的公司名称列表
	InvoiceTypes  []string                     `json:"invoice_types"` // 允许的发票类型列表
	ApplyDate     time.Time                    `json:"apply_date"`    // 报销申请日期
}

// InvoiceValidator 发票校验器接口
type InvoiceValidator interface {
	// ValidateSingle 校验单个发票
	ValidateSingle(ctx context.Context, req *InvoiceValidationRequest) (*InvoiceValidationResult, error)

	// ValidateBatch 批量校验发票
	ValidateBatch(ctx context.Context, reqs []*InvoiceValidationRequest) ([]*InvoiceValidationResult, error)

	// LoadRules 加载校验规则
	LoadRules(ctx context.Context) error

	// GetRuleDefinitions 获取规则定义
	GetRuleDefinitions() []*RuleDefinition
}

// RuleDefinition 规则定义
type RuleDefinition struct {
	ID          string `json:"id"`          // 规则ID
	RuleCode    string `json:"rule_code"`   // 规则编码
	Name        string `json:"name"`        // 规则名称
	Type        string `json:"type"`        // 规则类型
	Category    string `json:"category"`    // 规则分类
	Description string `json:"description"` // 规则描述
	Definition  string `json:"definition"`  // 规则定义(Grule语法)
	Priority    int    `json:"priority"`    // 优先级
	Enabled     bool   `json:"enabled"`     // 是否启用
}

// InvoiceValidatorImpl 发票校验器实现
type InvoiceValidatorImpl struct {
	ruleEngine *GRuleEngine
	repository Repository
	logger     logger.Logger
	rules      []*RuleDefinition
}

// NewInvoiceValidator 创建发票校验器
func NewInvoiceValidator(engine *GRuleEngine, repo Repository, log logger.Logger) InvoiceValidator {
	return &InvoiceValidatorImpl{
		ruleEngine: engine,
		repository: repo,
		logger:     log,
		rules:      make([]*RuleDefinition, 0),
	}
}

// ValidateSingle 校验单个发票
func (v *InvoiceValidatorImpl) ValidateSingle(ctx context.Context, req *InvoiceValidationRequest) (*InvoiceValidationResult, error) {
	if req == nil || req.Invoice == nil {
		v.logger.WithContext(ctx).Error("发票校验请求为空")
		return nil, errors.New("发票校验请求为空")
	}

	v.logger.WithContext(ctx).Info("开始校验发票",
		logger.NewField("发票ID", req.Invoice.ID),
		logger.NewField("发票号码", req.Invoice.Number))

	// 创建校验结果
	result := &InvoiceValidationResult{
		Passed:     true,
		InvoiceID:  req.Invoice.ID,
		Violations: make([]*InvoiceViolation, 0),
		Timestamp:  time.Now(),
	}

	// 执行Grule规则引擎校验（包含所有刚性规则）
	if err := v.executeRulesWithPriority(ctx, req, result); err != nil {
		v.logger.WithContext(ctx).Error("执行规则校验失败",
			logger.NewField("发票ID", req.Invoice.ID),
			logger.NewField("error", err.Error()))
		return nil, err
	}

	// 生成校验结果摘要
	v.generateSummary(result)

	v.logger.WithContext(ctx).Info("发票校验完成",
		logger.NewField("发票ID", req.Invoice.ID),
		logger.NewField("校验结果", result.Passed),
		logger.NewField("违规数量", len(result.Violations)))

	return result, nil
}

// ValidateBatch 批量校验发票
func (v *InvoiceValidatorImpl) ValidateBatch(ctx context.Context, reqs []*InvoiceValidationRequest) ([]*InvoiceValidationResult, error) {
	if len(reqs) == 0 {
		return nil, errors.New("发票校验请求列表为空")
	}

	v.logger.WithContext(ctx).Info("开始批量校验发票",
		logger.NewField("发票数量", len(reqs)))

	results := make([]*InvoiceValidationResult, 0, len(reqs))
	for _, req := range reqs {
		result, err := v.ValidateSingle(ctx, req)
		if err != nil {
			v.logger.WithContext(ctx).Error("校验发票失败",
				logger.NewField("发票ID", req.Invoice.ID),
				logger.NewField("error", err.Error()))
			// 创建失败结果
			result = &InvoiceValidationResult{
				Passed:    false,
				InvoiceID: req.Invoice.ID,
				Violations: []*InvoiceViolation{
					{
						RuleID:   "system_error",
						RuleName: "系统错误",
						RuleType: "系统",
						Severity: "高",
						Message:  "发票校验过程中发生系统错误",
					},
				},
				Summary:   "发票校验失败",
				Timestamp: time.Now(),
			}
		}
		results = append(results, result)
	}

	v.logger.WithContext(ctx).Info("批量校验发票完成",
		logger.NewField("发票数量", len(reqs)),
		logger.NewField("通过数量", v.countPassed(results)))

	return results, nil
}

// LoadRules 加载校验规则
func (v *InvoiceValidatorImpl) LoadRules(ctx context.Context) error {
	v.logger.WithContext(ctx).Info("加载发票校验规则")

	// 初始化基础规则定义（从数据库加载）
	v.initBasicRules(ctx)

	// 将规则加载到规则引擎
	for _, ruleDef := range v.rules {
		if !ruleDef.Enabled {
			continue
		}

		rule := &Rule{
			ID:         ruleDef.ID,
			RuleCode:   ruleDef.RuleCode,
			Name:       ruleDef.Name,
			Type:       ruleDef.Type,
			Definition: ruleDef.Definition,
			Priority:   ruleDef.Priority,
			Enabled:    ruleDef.Enabled,
		}

		if err := v.ruleEngine.LoadRule(ctx, rule); err != nil {
			v.logger.WithContext(ctx).Error("加载规则到引擎失败",
				logger.NewField("规则ID", ruleDef.ID),
				logger.NewField("规则名称", ruleDef.Name),
				logger.NewField("error", err.Error()))
			// 继续加载其他规则
			continue
		}
	}

	v.logger.WithContext(ctx).Info("发票校验规则加载完成",
		logger.NewField("规则数量", len(v.rules)))

	return nil
}

// initBasicRules 初始化基础规则定义
func (v *InvoiceValidatorImpl) initBasicRules(ctx context.Context) {
	v.logger.WithContext(ctx).Info("初始化基础发票校验规则")

	// 从数据库加载所有启用的发票校验规则
	rules, err := v.loadRulesFromDatabase(ctx)
	if err != nil {
		v.logger.WithContext(ctx).Error("从数据库加载规则失败",
			logger.NewField("error", err.Error()))
		return
	}

	// 将加载的规则添加到规则列表中
	v.rules = append(v.rules, rules...)

	v.logger.WithContext(ctx).Info("基础发票校验规则初始化完成",
		logger.NewField("规则数量", len(rules)))
}

// loadRulesFromDatabase 从数据库加载发票校验规则
func (v *InvoiceValidatorImpl) loadRulesFromDatabase(ctx context.Context) ([]*RuleDefinition, error) {
	v.logger.WithContext(ctx).Info("从数据库加载发票校验规则")

	// 从数据库查询所有启用的规则
	filter := &RuleFilter{
		Category: "发票校验",
		Status:   "enabled",
		Size:     1000,
	}

	rules, _, err := v.repository.ListRules(ctx, filter)
	if err != nil {
		v.logger.WithContext(ctx).Error("查询发票校验规则失败",
			logger.NewField("error", err.Error()))
		return nil, err
	}

	// 转换为规则定义
	ruleDefinitions := make([]*RuleDefinition, 0, len(rules))
	for _, rule := range rules {
		ruleDef := &RuleDefinition{
			ID:          rule.ID,
			RuleCode:    rule.RuleCode,
			Name:        rule.Name,
			Type:        rule.Type,
			Category:    rule.Category,
			Description: rule.Description,
			Definition:  rule.Definition,
			Priority:    rule.Priority,
			Enabled:     rule.Enabled,
		}
		ruleDefinitions = append(ruleDefinitions, ruleDef)
	}

	v.logger.WithContext(ctx).Info("发票校验规则加载完成",
		logger.NewField("规则数量", len(ruleDefinitions)))

	return ruleDefinitions, nil
}

// GetRuleDefinitions 获取规则定义
func (v *InvoiceValidatorImpl) GetRuleDefinitions() []*RuleDefinition {
	return v.rules
}

// countPassed 统计通过的校验结果数量
func (v *InvoiceValidatorImpl) countPassed(results []*InvoiceValidationResult) int {
	count := 0
	for _, result := range results {
		if result.Passed {
			count++
		}
	}
	return count
}

// generateSummary 生成校验结果摘要
func (v *InvoiceValidatorImpl) generateSummary(result *InvoiceValidationResult) {
	if result.Passed {
		result.Summary = "发票校验通过"
		return
	}

	// 按严重程度统计违规
	highCount := 0
	mediumCount := 0
	lowCount := 0

	for _, violation := range result.Violations {
		switch violation.Severity {
		case "高":
			highCount++
		case "中":
			mediumCount++
		case "低":
			lowCount++
		}
	}

	result.Summary = "发票校验未通过"
	if highCount > 0 {
		result.Summary += "，存在" + string(rune(highCount+'0')) + "项高风险违规"
	}
	if mediumCount > 0 {
		result.Summary += "，存在" + string(rune(mediumCount+'0')) + "项中风险违规"
	}
	if lowCount > 0 {
		result.Summary += "，存在" + string(rune(lowCount+'0')) + "项低风险违规"
	}
}
