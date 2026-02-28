package ruleengine

import (
	"time"
)

type Rule struct {
	ID          string      `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name        string      `json:"name" gorm:"type:varchar(100);not null"`
	Description string      `json:"description" gorm:"type:text"`
	Conditions  []Condition `json:"conditions" gorm:"foreignKey:RuleID;references:ID"`
	Decision    Decision    `json:"decision" gorm:"foreignKey:RuleID;references:ID"`
	Priority    int         `json:"priority" gorm:"default:0"`
	Enabled     bool        `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Rule) TableName() string {
	return "rule_engine_rules"
}

type RuleValidationResult struct {
	RuleID        string                 `json:"rule_id"`
	RuleName      string                 `json:"rule_name"`
	RuleType      string                 `json:"rule_type"`
	Passed        bool                   `json:"passed"`
	Message       string                 `json:"message"`
	Details       string                 `json:"details"`
	Severity      string                 `json:"severity"`
	Priority      int                    `json:"priority"`
	ExecutionTime int64                  `json:"execution_time"`
	Data          map[string]interface{} `json:"data"`
	Timestamp     time.Time              `json:"timestamp"`
	Violations    []interface{}          `json:"violations"`
}

type Condition struct {
	ID        string `json:"id" gorm:"primaryKey;type:varchar(36)"`
	RuleID    string `json:"rule_id" gorm:"type:varchar(36);not null;index"`
	FeatureID string `json:"feature_id" gorm:"type:varchar(36);not null;index"`
	Operator  string `json:"operator" gorm:"type:varchar(20);not null"`
	Value     string `json:"value" gorm:"type:text"`
	LogicOp   string `json:"logic_op" gorm:"type:varchar(10);default:'and'"`
	SortOrder int    `json:"sort_order" gorm:"default:0"`
}

func (Condition) TableName() string {
	return "conditions"
}

type Decision struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	RuleID    string    `json:"rule_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	Type      string    `json:"type" gorm:"type:varchar(20);not null"`
	Reason    string    `json:"reason" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Decision) TableName() string {
	return "decisions"
}

type RuleFilter struct {
	Name      string `json:"name"`
	Enabled   *bool  `json:"enabled"`
	FeatureID string `json:"feature_id"`
	Page      int    `json:"page"`
	Size      int    `json:"size"`
}

type ConditionFilter struct {
	RuleID    string `json:"rule_id"`
	FeatureID string `json:"feature_id"`
}

type DecisionFilter struct {
	RuleID string `json:"rule_id"`
	Type   string `json:"type"`
}

const (
	OperatorEqual        = "eq"
	OperatorNotEqual     = "ne"
	OperatorGreaterThan  = "gt"
	OperatorGreaterEqual = "gte"
	OperatorLessThan     = "lt"
	OperatorLessEqual    = "lte"
	OperatorIn           = "in"
	OperatorNotIn        = "not_in"
	OperatorContains     = "contains"
	OperatorNotContains  = "not_contains"
)

const (
	LogicOpAnd = "and"
	LogicOpOr  = "or"
)

const (
	DecisionTypeApprove = "approve"
	DecisionTypeReject  = "reject"
	DecisionTypeMark    = "mark"
)

type RuleResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Conditions  []*ConditionResponse `json:"conditions"`
	Decision    DecisionResponse     `json:"decision"`
	Priority    int                  `json:"priority"`
	Enabled     bool                 `json:"enabled"`
	CreatedAt   string               `json:"created_at"`
	UpdatedAt   string               `json:"updated_at"`
}

type ConditionResponse struct {
	ID        string `json:"id"`
	FeatureID string `json:"feature_id"`
	Operator  string `json:"operator"`
	Value     string `json:"value"`
	LogicOp   string `json:"logic_op"`
}

type DecisionResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type FeatureResponse struct {
	ID             string                  `json:"id"`
	Name           string                  `json:"name"`
	Code           string                  `json:"code"`
	Description    string                  `json:"description"`
	Type           string                  `json:"type"`
	ValueType      string                  `json:"value_type"`
	Category       string                  `json:"category"`
	Enabled        bool                    `json:"enabled"`
	FunctionName   string                  `json:"function_name"`
	FunctionConfig map[string]interface{}  `json:"function_config"`
	Values         []*FeatureValueResponse `json:"values"`
	CreatedAt      string                  `json:"created_at"`
	UpdatedAt      string                  `json:"updated_at"`
}

type FeatureValueResponse struct {
	ID        string `json:"id"`
	FeatureID string `json:"feature_id"`
	Value     string `json:"value"`
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
	Enabled   bool   `json:"enabled"`
}
