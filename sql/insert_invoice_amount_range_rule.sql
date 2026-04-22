-- 插入发票金额范围校验规则
-- 规则：发票金额必须在合理范围内（0.01元 - 100000元）

-- 插入特征
INSERT INTO features (id, name, code, type, value_type, category, function_name, function_config, description)
VALUES (
    'feat-invoice-amount-range',
    '发票金额范围',
    'invoice_amount_range',
    'boolean',
    'boolean',
    'validation',
    'invoice_amount_range',
    '{"min_amount": 0.01, "max_amount": 100000.00}',
    '检查发票金额是否在合理范围内（0.01元 - 100000元）'
)
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    type = VALUES(type),
    category = VALUES(category),
    function_name = VALUES(function_name),
    function_config = VALUES(function_config),
    description = VALUES(description),
    updated_at = NOW();

-- 插入规则
INSERT INTO rule_engine_rules (id, name, description, priority, enabled)
VALUES (
    'rule-invoice-amount-range',
    '发票金额范围校验规则',
    '检查发票金额是否在合理范围内（0.01元 - 100000元）',
    70,
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
    'cond-invoice-amount-range',
    'rule-invoice-amount-range',
    'feat-invoice-amount-range',
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
    'decision-invoice-amount-range',
    'rule-invoice-amount-range',
    'reject',
    '发票金额不在合理范围内（0.01元 - 100000元），不符合报销要求',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON DUPLICATE KEY UPDATE
    rule_id = VALUES(rule_id),
    type = VALUES(type),
    reason = VALUES(reason),
    updated_at = CURRENT_TIMESTAMP;
