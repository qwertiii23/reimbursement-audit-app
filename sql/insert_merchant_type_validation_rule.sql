-- 插入商户类型校验规则
-- 规则：商户类型必须在允许的范围内（一般纳税人、小规模纳税人、个体工商户）

-- 插入特征
INSERT INTO features (id, name, code, type, value_type, category, function_name, function_config, description)
VALUES (
    'feat-merchant-type-validation',
    '商户类型校验',
    'merchant_type_validation',
    'boolean',
    'boolean',
    'validation',
    'merchant_type_validation',
    '{"allowed_types": ["一般纳税人", "小规模纳税人", "个体工商户"]}',
    '检查商户类型是否在允许的范围内（一般纳税人、小规模纳税人、个体工商户）'
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
    'rule-merchant-type-validation',
    '商户类型校验规则',
    '检查商户类型是否在允许的范围内（一般纳税人、小规模纳税人、个体工商户）',
    72,
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
    'cond-merchant-type-validation',
    'rule-merchant-type-validation',
    'feat-merchant-type-validation',
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
    'decision-merchant-type-validation',
    'rule-merchant-type-validation',
    'reject',
    '商户类型不在允许的范围内（一般纳税人、小规模纳税人、个体工商户），不符合报销要求',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON DUPLICATE KEY UPDATE
    rule_id = VALUES(rule_id),
    type = VALUES(type),
    reason = VALUES(reason),
    updated_at = CURRENT_TIMESTAMP;
