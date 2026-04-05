-- 插入餐饮费校验特征
INSERT INTO features (id, name, code, type, value_type, category, function_name, function_config, description)
VALUES (
    'feat-meal',
    '餐饮费校验',
    'meal_validation',
    'boolean',
    'boolean',
    'validation',
    'meal_validation',
    '{"enable_daily_limit": true, "enable_per_meal_limit": true}',
    '根据城市等级和员工职级校验餐饮费'
)
ON DUPLICATE KEY UPDATE name=name;

-- 插入餐饮费校验规则
INSERT INTO rule_engine_rules (id, name, description, priority, enabled)
VALUES (
    'rule-meal',
    '餐饮费校验规则',
    '根据城市等级和员工职级校验餐饮费是否在标准范围内',
    84,
    1
)
ON DUPLICATE KEY UPDATE name=name;

-- 插入条件：餐饮费校验特征为true
INSERT INTO conditions (id, rule_id, feature_id, operator, value)
VALUES (
    'cond-meal-1',
    'rule-meal',
    'feat-meal',
    '==',
    'true'
)
ON DUPLICATE KEY UPDATE value=value;

-- 插入决策：通过
INSERT INTO decisions (id, rule_id, type, reason, action)
VALUES (
    'dec-meal-1',
    'rule-meal',
    'approval',
    '餐饮费符合标准',
    'approve'
)
ON DUPLICATE KEY UPDATE reason=reason;

-- 插入决策：拒绝
INSERT INTO decisions (id, rule_id, type, reason, action)
VALUES (
    'dec-meal-2',
    'rule-meal',
    'rejection',
    '餐饮费超出标准',
    'reject'
)
ON DUPLICATE KEY UPDATE reason=reason;
