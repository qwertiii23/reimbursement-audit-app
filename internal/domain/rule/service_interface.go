package rule

import (
	"context"

	"reimbursement-audit/internal/api/request"
)

// RuleServiceInterface 规则服务接口
type RuleServiceInterface interface {
	CreateRule(ctx context.Context, req *request.CreateRuleRequest) (*Rule, error)
	UpdateRule(ctx context.Context, req *request.UpdateRuleRequest) (*Rule, error)
	DeleteRule(ctx context.Context, ruleID string) error
	GetRules(ctx context.Context, filter *RuleFilter) ([]*Rule, int64, error)
	GetRuleByID(ctx context.Context, ruleID string) (*Rule, error)
	EnableRule(ctx context.Context, ruleID string) error
	DisableRule(ctx context.Context, ruleID string) error
	TestRule(ctx context.Context, rule *Rule, testData interface{}) (*RuleValidationResult, error)
}
