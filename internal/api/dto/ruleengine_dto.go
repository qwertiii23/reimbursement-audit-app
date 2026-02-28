package ruleengine

type CreateRuleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Conditions  []ConditionRequest `json:"conditions" binding:"required"`
	Decision    DecisionRequest `json:"decision" binding:"required"`
	Priority    int `json:"priority"`
	Enabled     bool `json:"enabled"`
}

type UpdateRuleRequest struct {
	ID          string `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Conditions  []ConditionRequest `json:"conditions" binding:"required"`
	Decision    DecisionRequest `json:"decision" binding:"required"`
	Priority    int `json:"priority"`
	Enabled     bool `json:"enabled"`
}

type ConditionRequest struct {
	FeatureID string `json:"feature_id" binding:"required"`
	Operator  string `json:"operator" binding:"required"`
	Value     string `json:"value"`
	ValueList []string `json:"value_list"`
	LogicOp   string `json:"logic_op"`
	SortOrder int `json:"sort_order"`
}

type DecisionRequest struct {
	Type   string `json:"type" binding:"required"`
	Reason string `json:"reason"`
}

type RuleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Conditions  []ConditionResponse `json:"conditions"`
	Decision    DecisionResponse `json:"decision"`
	Priority    int `json:"priority"`
	Enabled     bool `json:"enabled"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ConditionResponse struct {
	ID        string `json:"id"`
	RuleID    string `json:"rule_id"`
	FeatureID string `json:"feature_id"`
	Operator  string `json:"operator"`
	Value     string `json:"value"`
	ValueList []string `json:"value_list"`
	LogicOp   string `json:"logic_op"`
	SortOrder int `json:"sort_order"`
}

type DecisionResponse struct {
	ID        string `json:"id"`
	RuleID    string `json:"rule_id"`
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type RuleListResponse struct {
	Rules []*RuleResponse `json:"rules"`
	Total int64            `json:"total"`
}

type CreateFeatureRequest struct {
	Name      string `json:"name" binding:"required"`
	Code      string `json:"code" binding:"required"`
	Description string `json:"description"`
	Type      string `json:"type" binding:"required"`
	ValueType string `json:"value_type" binding:"required"`
	Category  string `json:"category"`
	Enabled   bool `json:"enabled"`
	Values    []FeatureValueRequest `json:"values"`
}

type UpdateFeatureRequest struct {
	ID        string `json:"id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Code      string `json:"code" binding:"required"`
	Description string `json:"description"`
	Type      string `json:"type" binding:"required"`
	ValueType string `json:"value_type" binding:"required"`
	Category  string `json:"category"`
	Enabled   bool `json:"enabled"`
	Values    []FeatureValueRequest `json:"values"`
}

type FeatureValueRequest struct {
	Value     string `json:"value" binding:"required"`
	Label     string `json:"label" binding:"required"`
	SortOrder int `json:"sort_order"`
	Enabled   bool `json:"enabled"`
}

type FeatureResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Type        string `json:"type"`
	ValueType   string `json:"value_type"`
	Category    string `json:"category"`
	Enabled     bool `json:"enabled"`
	Values      []FeatureValueResponse `json:"values"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type FeatureValueResponse struct {
	ID        string `json:"id"`
	FeatureID string `json:"feature_id"`
	Value     string `json:"value"`
	Label     string `json:"label"`
	SortOrder int `json:"sort_order"`
	Enabled   bool `json:"enabled"`
}

type FeatureListResponse struct {
	Features []*FeatureResponse `json:"features"`
	Total    int64              `json:"total"`
}

type EvaluateRuleRequest struct {
	RuleID string `json:"rule_id" binding:"required"`
	Data   map[string]interface{} `json:"data" binding:"required"`
}

type EvaluateRuleResponse struct {
	RuleID         string `json:"rule_id"`
	RuleName       string `json:"rule_name"`
	Passed          bool `json:"passed"`
	Message         string `json:"message"`
	DecisionType    string `json:"decision_type"`
	DecisionReason  string `json:"decision_reason"`
}
