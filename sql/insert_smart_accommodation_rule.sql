-- 插入智能住宿费校验规则

-- 插入特征
INSERT INTO features (id, name, code, type, value_type, category, function_name, function_config, description)
VALUES (
    'feat-smart-accommodation',
    '智能住宿费校验',
    'smart_accommodation_validation',
    'boolean',
    'boolean',
    'validation',
    'smart_accommodation_validation',
    '{"enable_event_adjustment": true, "enable_holiday_adjustment": true}',
    '智能住宿费校验，根据城市等级、活动、节假日动态调整标准'
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
    'rule-smart-accommodation',
    '智能住宿费校验规则',
    '根据城市等级、活动、节假日动态调整住宿费标准，智能审核住宿费报销',
    90,
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
    'cond-smart-accommodation',
    'rule-smart-accommodation',
    'feat-smart-accommodation',
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
    'decision-smart-accommodation',
    'rule-smart-accommodation',
    'review',
    '住宿费超出智能调整后的标准，需要人工审核',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON DUPLICATE KEY UPDATE
    rule_id = VALUES(rule_id),
    type = VALUES(type),
    reason = VALUES(reason),
    updated_at = CURRENT_TIMESTAMP;
