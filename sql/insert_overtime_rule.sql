-- 插入加班费校验特征
INSERT INTO features (id, name, code, type, value_type, category, function_name, function_config, description)
VALUES (
    'feat-overtime',
    '加班费校验',
    'overtime_validation',
    'boolean',
    'boolean',
    'validation',
    'overtime_validation',
    '{"enable_daily_limit": true, "enable_monthly_limit": true}',
    '根据加班时长和员工职级校验加班费'
)
ON DUPLICATE KEY UPDATE name=name;

-- 插入加班费校验规则
INSERT INTO rule_engine_rules (id, name, description, priority, enabled)
VALUES (
    'rule-overtime',
    '加班费校验规则',
    '根据加班时长和员工职级校验加班费是否在标准范围内',
    81,
    1
)
ON DUPLICATE KEY UPDATE name=name;

-- 插入条件：加班费校验特征为true
INSERT INTO conditions (id, rule_id, feature_id, operator, value)
VALUES (
    'cond-overtime-1',
    'rule-overtime',
    'feat-overtime',
    '==',
    'true'
)
ON DUPLICATE KEY UPDATE value=value;

-- 插入决策：通过
INSERT INTO decisions (id, rule_id, type, reason, action)
VALUES (
    'dec-overtime-1',
    'rule-overtime',
    'approval',
    '加班费符合标准',
    'approve'
)
ON DUPLICATE KEY UPDATE reason=reason;

-- 插入决策：拒绝
INSERT INTO decisions (id, rule_id, type, reason, action)
VALUES (
    'dec-overtime-2',
    'rule-overtime',
    'rejection',
    '加班费超出标准',
    'reject'
)
ON DUPLICATE KEY UPDATE reason=reason;
