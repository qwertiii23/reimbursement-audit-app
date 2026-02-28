package ruleengine

import "context"

type RuleRepository interface {
	CreateRule(ctx context.Context, rule *Rule) error
	GetRuleByID(ctx context.Context, id string) (*Rule, error)
	UpdateRule(ctx context.Context, rule *Rule) error
	DeleteRule(ctx context.Context, id string) error
	ListRules(ctx context.Context, filter *RuleFilter) ([]*Rule, int64, error)
	EnableRule(ctx context.Context, id string) error
	DisableRule(ctx context.Context, id string) error
}

type ConditionRepository interface {
	CreateCondition(ctx context.Context, condition *Condition) error
	GetConditionByID(ctx context.Context, id string) (*Condition, error)
	UpdateCondition(ctx context.Context, condition *Condition) error
	DeleteCondition(ctx context.Context, id string) error
	ListConditions(ctx context.Context, filter *ConditionFilter) ([]*Condition, error)
	DeleteConditionsByRuleID(ctx context.Context, ruleID string) error
}

type DecisionRepository interface {
	CreateDecision(ctx context.Context, decision *Decision) error
	GetDecisionByRuleID(ctx context.Context, ruleID string) (*Decision, error)
	UpdateDecision(ctx context.Context, decision *Decision) error
	DeleteDecision(ctx context.Context, id string) error
}

type FeatureRepository interface {
	CreateFeature(ctx context.Context, feature *Feature) error
	GetFeatureByID(ctx context.Context, id string) (*Feature, error)
	GetFeatureByCode(ctx context.Context, code string) (*Feature, error)
	UpdateFeature(ctx context.Context, feature *Feature) error
	DeleteFeature(ctx context.Context, id string) error
	ListFeatures(ctx context.Context, filter *FeatureFilter) ([]*Feature, int64, error)
	EnableFeature(ctx context.Context, id string) error
	DisableFeature(ctx context.Context, id string) error
}

type FeatureValueRepository interface {
	CreateFeatureValue(ctx context.Context, value *FeatureValue) error
	GetFeatureValueByID(ctx context.Context, id string) (*FeatureValue, error)
	UpdateFeatureValue(ctx context.Context, value *FeatureValue) error
	DeleteFeatureValue(ctx context.Context, id string) error
	ListFeatureValues(ctx context.Context, filter *FeatureValueFilter) ([]*FeatureValue, error)
	DeleteFeatureValuesByFeatureID(ctx context.Context, featureID string) error
}
