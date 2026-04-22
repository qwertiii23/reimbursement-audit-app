-- 插入交通费校验特征
INSERT INTO features (id, name, code, type, value_type, category, function_name, function_config, description)
VALUES (
    'feat-transportation',
    '交通费校验',
    'transportation_validation',
    'boolean',
    'boolean',
    'validation',
    'transportation_validation',
    '{"enable_daily_limit": true, "enable_trip_limit": true}',
    '根据出行方式和员工职级校验交通费'
)
ON DUPLICATE KEY UPDATE name=name;

-- 插入交通费校验规则
INSERT INTO rule_engine_rules (id, name, description, priority, enabled)
VALUES (
    'rule-transportation',
    '交通费校验规则',
    '根据出行方式和员工职级校验交通费是否在标准范围内',
    85,
    1
)
ON DUPLICATE KEY UPDATE name=name;

-- 插入条件：交通费校验特征为true
INSERT INTO conditions (id, rule_id, feature_id, operator, value)
VALUES (
    'cond-transportation-1',
    'rule-transportation',
    'feat-transportation',
    '==',
    'true'
)
ON DUPLICATE KEY UPDATE value=value;

-- 插入决策：通过
INSERT INTO decisions (id, rule_id, type, reason, action)
VALUES (
    'dec-transportation-1',
    'rule-transportation',
    'approval',
    '交通费符合标准',
    'approve'
)
ON DUPLICATE KEY UPDATE reason=reason;

-- 插入决策：拒绝
INSERT INTO decisions (id, rule_id, type, reason, action)
VALUES (
    'dec-transportation-2',
    'rule-transportation',
    'rejection',
    '交通费超出标准',
    'reject'
)
ON DUPLICATE KEY UPDATE reason=reason;
