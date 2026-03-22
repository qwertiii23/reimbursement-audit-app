package ruleengine

import (
	"fmt"
	"strings"
)

type ExpressionParser struct {
	featureMap map[string]*Feature
}

func NewExpressionParser(features []*Feature) *ExpressionParser {
	featureMap := make(map[string]*Feature)
	for _, feature := range features {
		featureMap[feature.Code] = feature
	}
	return &ExpressionParser{
		featureMap: featureMap,
	}
}

type ExpressionNode interface {
	Evaluate(data map[string]interface{}) (bool, error)
	String() string
}

type ConditionNode struct {
	FeatureCode string
	Operator    string
	Value       string
}

func (n *ConditionNode) Evaluate(data map[string]interface{}) (bool, error) {
	value, exists := data[n.FeatureCode]
	if !exists {
		return false, fmt.Errorf("特征 %s 不存在", n.FeatureCode)
	}

	switch n.Operator {
	case OperatorEqual:
		return fmt.Sprintf("%v", value) == n.Value, nil
	case OperatorNotEqual:
		return fmt.Sprintf("%v", value) != n.Value, nil
	case OperatorGreaterThan:
		return compareNumbers(value, n.Value, ">")
	case OperatorGreaterEqual:
		return compareNumbers(value, n.Value, ">=")
	case OperatorLessThan:
		return compareNumbers(value, n.Value, "<")
	case OperatorLessEqual:
		return compareNumbers(value, n.Value, "<=")
	case OperatorContains:
		return containsString(value, n.Value, true)
	case OperatorNotContains:
		return containsString(value, n.Value, false)
	default:
		return false, fmt.Errorf("不支持的运算符: %s", n.Operator)
	}
}

func (n *ConditionNode) String() string {
	return fmt.Sprintf("%s %s %s", n.FeatureCode, n.Operator, n.Value)
}

type LogicalNode struct {
	Left     ExpressionNode
	Operator string
	Right    ExpressionNode
}

func (n *LogicalNode) Evaluate(data map[string]interface{}) (bool, error) {
	leftResult, err := n.Left.Evaluate(data)
	if err != nil {
		return false, err
	}

	if n.Operator == LogicOpAnd && !leftResult {
		return false, nil
	}

	if n.Operator == LogicOpOr && leftResult {
		return true, nil
	}

	rightResult, err := n.Right.Evaluate(data)
	if err != nil {
		return false, err
	}

	if n.Operator == LogicOpAnd {
		return leftResult && rightResult, nil
	}
	return leftResult || rightResult, nil
}

func (n *LogicalNode) String() string {
	return fmt.Sprintf("(%s %s %s)", n.Left.String(), n.Operator, n.Right.String())
}

type GroupNode struct {
	Expression ExpressionNode
}

func (n *GroupNode) Evaluate(data map[string]interface{}) (bool, error) {
	return n.Expression.Evaluate(data)
}

func (n *GroupNode) String() string {
	return fmt.Sprintf("(%s)", n.Expression.String())
}

func (p *ExpressionParser) Parse(expression string) (ExpressionNode, error) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil, fmt.Errorf("表达式不能为空")
	}

	if strings.HasPrefix(expression, "(") && strings.HasSuffix(expression, ")") {
		inner := strings.TrimSpace(expression[1 : len(expression)-1])
		node, err := p.Parse(inner)
		if err != nil {
			return nil, err
		}
		return &GroupNode{Expression: node}, nil
	}

	return p.parseLogicalExpression(expression)
}

func (p *ExpressionParser) parseLogicalExpression(expression string) (ExpressionNode, error) {
	operators := []string{LogicOpOr, LogicOpAnd}

	for _, op := range operators {
		if strings.Contains(expression, op) {
			parts := strings.Split(expression, op)
			if len(parts) != 2 {
				continue
			}

			left, err := p.parseLogicalExpression(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, err
			}

			right, err := p.parseLogicalExpression(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, err
			}

			return &LogicalNode{
				Left:     left,
				Operator: op,
				Right:    right,
			}, nil
		}
	}

	return p.parseCondition(expression)
}

func (p *ExpressionParser) parseCondition(expression string) (ExpressionNode, error) {
	operators := []string{
		OperatorNotContains, OperatorContains,
		OperatorGreaterEqual, OperatorLessEqual,
		OperatorGreaterThan, OperatorLessThan,
		OperatorNotEqual, OperatorEqual,
	}

	for _, op := range operators {
		if strings.Contains(expression, op) {
			parts := strings.SplitN(expression, op, 2)
			if len(parts) != 2 {
				continue
			}

			featureCode := strings.TrimSpace(parts[0])
			valueStr := strings.TrimSpace(parts[1])

			_, exists := p.featureMap[featureCode]
			if !exists {
				return nil, fmt.Errorf("特征 %s 不存在", featureCode)
			}

			return &ConditionNode{
				FeatureCode: featureCode,
				Operator:    op,
				Value:       valueStr,
			}, nil
		}
	}

	return nil, fmt.Errorf("无法解析条件: %s", expression)
}

func (p *ExpressionParser) ParseRule(rule *Rule) (ExpressionNode, error) {
	if len(rule.Conditions) == 0 {
		return nil, fmt.Errorf("规则没有条件")
	}

	if len(rule.Conditions) == 1 {
		return p.parseConditionFromModel(&rule.Conditions[0])
	}

	var node ExpressionNode
	var err error

	node, err = p.parseConditionFromModel(&rule.Conditions[0])
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(rule.Conditions); i++ {
		conditionNode, err := p.parseConditionFromModel(&rule.Conditions[i])
		if err != nil {
			return nil, err
		}

		logicOp := rule.Conditions[i].LogicOp
		if logicOp == "" {
			logicOp = LogicOpAnd
		}

		node = &LogicalNode{
			Left:     node,
			Operator: logicOp,
			Right:    conditionNode,
		}
	}

	return node, nil
}

func (p *ExpressionParser) parseConditionFromModel(condition *Condition) (ExpressionNode, error) {
	_, exists := p.featureMap[condition.FeatureID]
	if !exists {
		return nil, fmt.Errorf("特征 %s 不存在", condition.FeatureID)
	}

	return &ConditionNode{
		FeatureCode: condition.FeatureID,
		Operator:    condition.Operator,
		Value:       condition.Value,
	}, nil
}

func compareNumbers(value interface{}, target string, operator string) (bool, error) {
	valueStr := fmt.Sprintf("%v", value)
	var num1, num2 float64
	_, err1 := fmt.Sscanf(valueStr, "%f", &num1)
	_, err2 := fmt.Sscanf(target, "%f", &num2)

	if err1 != nil || err2 != nil {
		return false, fmt.Errorf("无法比较非数字值")
	}

	switch operator {
	case ">":
		return num1 > num2, nil
	case ">=":
		return num1 >= num2, nil
	case "<":
		return num1 < num2, nil
	case "<=":
		return num1 <= num2, nil
	default:
		return false, fmt.Errorf("不支持的比较运算符: %s", operator)
	}
}

func containsString(value interface{}, target string, shouldContain bool) (bool, error) {
	valueStr := fmt.Sprintf("%v", value)
	if shouldContain {
		return strings.Contains(valueStr, target), nil
	}
	return !strings.Contains(valueStr, target), nil
}

func ValidateExpression(expression string) error {
	if expression == "" {
		return fmt.Errorf("表达式不能为空")
	}

	balance := 0
	for _, char := range expression {
		if char == '(' {
			balance++
		} else if char == ')' {
			balance--
			if balance < 0 {
				return fmt.Errorf("括号不匹配")
			}
		}
	}

	if balance != 0 {
		return fmt.Errorf("括号不匹配")
	}

	operators := []string{
		OperatorNotContains, OperatorContains,
		OperatorGreaterEqual, OperatorLessEqual,
		OperatorGreaterThan, OperatorLessThan,
		OperatorNotEqual, OperatorEqual,
		LogicOpAnd, LogicOpOr,
	}

	hasOperator := false
	for _, op := range operators {
		if strings.Contains(expression, op) {
			hasOperator = true
			break
		}
	}

	if !hasOperator {
		return fmt.Errorf("表达式必须包含运算符")
	}

	return nil
}

func FormatExpression(rule *Rule) string {
	if len(rule.Conditions) == 0 {
		return ""
	}

	var parts []string
	for i, cond := range rule.Conditions {
		if i > 0 {
			logicOp := cond.LogicOp
			if logicOp == "" {
				logicOp = LogicOpAnd
			}
			parts = append(parts, strings.ToUpper(logicOp))
		}

		parts = append(parts, fmt.Sprintf("%s %s %s", cond.FeatureID, cond.Operator, cond.Value))
	}

	return strings.Join(parts, " ")
}

func ParseRuleFromText(text string, featureMap map[string]*Feature) (*Rule, error) {
	rule := &Rule{
		ID:          "",
		Name:        "",
		Description: "",
		Priority:    0,
		Enabled:     true,
	}

	parser := NewExpressionParser([]*Feature{})
	for _, feature := range featureMap {
		parser.featureMap[feature.Code] = feature
	}

	node, err := parser.Parse(text)
	if err != nil {
		return nil, err
	}

	extractConditions(node, rule)

	return rule, nil
}

func extractConditions(node ExpressionNode, rule *Rule) {
	switch n := node.(type) {
	case *ConditionNode:
		rule.Conditions = append(rule.Conditions, Condition{
			FeatureID: n.FeatureCode,
			Operator:  n.Operator,
			Value:     n.Value,
			LogicOp:   LogicOpAnd,
		})
	case *LogicalNode:
		extractConditions(n.Left, rule)
		extractConditions(n.Right, rule)
		if len(rule.Conditions) > 0 {
			rule.Conditions[len(rule.Conditions)-1].LogicOp = n.Operator
		}
	case *GroupNode:
		extractConditions(n.Expression, rule)
	}
}

func OptimizeExpression(node ExpressionNode) ExpressionNode {
	switch n := node.(type) {
	case *LogicalNode:
		n.Left = OptimizeExpression(n.Left)
		n.Right = OptimizeExpression(n.Right)

		if leftCond, ok := n.Left.(*ConditionNode); ok {
			if rightCond, ok := n.Right.(*ConditionNode); ok {
				if leftCond.FeatureCode == rightCond.FeatureCode &&
					leftCond.Operator == rightCond.Operator &&
					leftCond.Value == rightCond.Value {
					return n.Left
				}
			}
		}

		return n
	case *GroupNode:
		n.Expression = OptimizeExpression(n.Expression)

		if _, ok := n.Expression.(*ConditionNode); ok {
			return n.Expression
		}

		return n
	default:
		return node
	}
}

func GetExpressionVariables(node ExpressionNode) []string {
	var vars []string
	var visited = make(map[string]bool)

	var traverse func(ExpressionNode)
	traverse = func(n ExpressionNode) {
		switch node := n.(type) {
		case *ConditionNode:
			if !visited[node.FeatureCode] {
				visited[node.FeatureCode] = true
				vars = append(vars, node.FeatureCode)
			}
		case *LogicalNode:
			traverse(node.Left)
			traverse(node.Right)
		case *GroupNode:
			traverse(node.Expression)
		}
	}

	traverse(node)
	return vars
}

func TestExpression(expression string, testData map[string]interface{}) (bool, error) {
	if err := ValidateExpression(expression); err != nil {
		return false, err
	}

	parser := NewExpressionParser([]*Feature{})
	node, err := parser.Parse(expression)
	if err != nil {
		return false, err
	}

	return node.Evaluate(testData)
}

func GenerateExpressionFromConditions(conditions []Condition) string {
	if len(conditions) == 0 {
		return ""
	}

	var parts []string
	for i, cond := range conditions {
		if i > 0 {
			logicOp := cond.LogicOp
			if logicOp == "" {
				logicOp = LogicOpAnd
			}
			parts = append(parts, strings.ToUpper(logicOp))
		}

		parts = append(parts, fmt.Sprintf("%s %s %s", cond.FeatureID, cond.Operator, cond.Value))
	}

	return strings.Join(parts, " ")
}
