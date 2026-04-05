-- 插入招待费校验特征
INSERT INTO features (id, name, code, type, value_type, category, function_name, function_config, description)
VALUES (
    'feat-entertainment',
    '招待费校验',
    'entertainment_validation',
    'boolean',
    'boolean',
    'validation',
    'entertainment_validation',
    '{"enable_daily_limit": true, "enable_per_person_limit": true}',
    '根据招待对象和员工职级校验招待费'
)
ON DUPLICATE KEY UPDATE name=name;

-- 插入招待费校验规则
INSERT INTO rule_engine_rules (id, name, description, priority, enabled)
VALUES (
    'rule-entertainment',
    '招待费校验规则',
    '根据招待对象和员工职级校验招待费是否在标准范围内',
    82,
    1
)
ON DUPLICATE KEY UPDATE name=name;

-- 插入条件：招待费校验特征为true
INSERT INTO conditions (id, rule_id, feature_id, operator, value)
VALUES (
    'cond-entertainment-1',
    'rule-entertainment',
    'feat-entertainment',
    '==',
    'true'
)
ON DUPLICATE KEY UPDATE value=value;

-- 插入决策：通过
INSERT INTO decisions (id, rule_id, type, reason, action)
VALUES (
    'dec-entertainment-1',
    'rule-entertainment',
    'approval',
    '招待费符合标准',
    'approve'
)
ON DUPLICATE KEY UPDATE reason=reason;

-- 插入决策：拒绝
INSERT INTO decisions (id, rule_id, type, reason, action)
VALUES (
    'dec-entertainment-2',
    'rule-entertainment',
    'rejection',
    '招待费超出标准',
    'reject'
)
ON DUPLICATE KEY UPDATE reason=reason;
