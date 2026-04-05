-- 插入发票代码/号码校验规则
-- 规则：发票代码和号码格式必须正确（代码12位数字，号码8位数字）

-- 插入特征
INSERT INTO features (id, name, code, type, value_type, category, function_name, function_config, description)
VALUES (
    'feat-invoice-code-number-validation',
    '发票代码号码校验',
    'invoice_code_number_validation',
    'boolean',
    'boolean',
    'validation',
    'invoice_code_number_validation',
    '{"code_pattern": "^\\\\d{12}$", "number_pattern": "^\\\\d{8}$"}',
    '检查发票代码和号码格式是否正确（代码12位数字，号码8位数字）'
)
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    type = VALUES(type),
    value_type = VALUES(value_type),
    category = VALUES(category),
    function_name = VALUES(function_name),
    function_config = VALUES(function_config),
    description = VALUES(description);

-- 插入规则
INSERT INTO rule_engine_rules (id, name, description, priority, enabled)
VALUES (
    'rule-invoice-code-number-validation',
    '发票代码号码校验规则',
    '检查发票代码和号码格式是否正确（代码12位数字，号码8位数字）',
    85,
    true
)
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    description = VALUES(description),
    priority = VALUES(priority),
    enabled = VALUES(enabled);

-- 插入条件
INSERT INTO conditions (id, rule_id, feature_id, operator, value, logic_op, sort_order)
VALUES (
    'cond-invoice-code-number-validation',
    'rule-invoice-code-number-validation',
    'feat-invoice-code-number-validation',
    'eq',
    'false',
    'and',
    1
)
ON DUPLICATE KEY UPDATE
    rule_id = VALUES(rule_id),
    feature_id = VALUES(feature_id),
    operator = VALUES(operator),
    value = VALUES(value),
    logic_op = VALUES(logic_op),
    sort_order = VALUES(sort_order);

-- 插入决策
INSERT INTO decisions (id, rule_id, type, reason, created_at, updated_at)
VALUES (
    'decision-invoice-code-num',
    'rule-invoice-code-number-validation',
    'reject',
    '发票代码或号码格式不正确（代码12位数字，号码8位数字），不符合报销要求',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON DUPLICATE KEY UPDATE
    rule_id = VALUES(rule_id),
    type = VALUES(type),
    reason = VALUES(reason),
    updated_at = CURRENT_TIMESTAMP;
