// repository.go 规则仓储接口
// 功能点：
// 1. 定义规则仓储接口
// 2. 提供规则CRUD操作抽象
// 3. 提供规则查询和筛选功能

package rule

import "context"

// Repository 规则仓储接口
type Repository interface {
	// CreateRule 创建规则
	CreateRule(ctx context.Context, rule *Rule) error

	// GetRuleByID 根据ID获取规则
	GetRuleByID(ctx context.Context, id string) (*Rule, error)

	// GetRuleByCode 根据规则编码获取规则
	GetRuleByCode(ctx context.Context, ruleCode string) (*Rule, error)

	// UpdateRule 更新规则
	UpdateRule(ctx context.Context, rule *Rule) error

	// DeleteRule 删除规则
	DeleteRule(ctx context.Context, id string) error

	// ListRules 根据过滤条件查询规则列表
	ListRules(ctx context.Context, filter *RuleFilter) ([]*Rule, int64, error)

	// CountRules 根据过滤条件统计规则数量
	CountRules(ctx context.Context, filter *RuleFilter) (int64, error)

	// EnableRule 启用规则
	EnableRule(ctx context.Context, id string) error

	// DisableRule 禁用规则
	DisableRule(ctx context.Context, id string) error

	// CheckRuleCodeExists 检查规则编码是否存在
	CheckRuleCodeExists(ctx context.Context, ruleCode string, excludeID string) (bool, error)
}
