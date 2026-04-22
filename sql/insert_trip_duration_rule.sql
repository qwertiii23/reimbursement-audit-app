-- 插入差旅天数校验特征
INSERT INTO features (id, name, code, type, value_type, category, function_name, function_config, description)
VALUES (
    'feat-trip-duration',
    '差旅天数校验',
    'trip_duration_validation',
    'boolean',
    'boolean',
    'validation',
    'trip_duration_validation',
    '{"max_meeting_days": 3, "max_training_days": 7, "max_project_days": 15, "max_visit_days": 5, "max_other_days": 10}',
    '检查出差天数是否合理'
)
ON DUPLICATE KEY UPDATE name=name;

-- 插入差旅天数校验规则
INSERT INTO rule_engine_rules (id, name, description, priority, enabled)
VALUES (
    'rule-trip-duration',
    '差旅天数校验规则',
    '检查出差天数是否合理',
    83,
    1
)
ON DUPLICATE KEY UPDATE name=name;

-- 插入条件：差旅天数有效
INSERT INTO conditions (id, rule_id, feature_id, operator, value)
VALUES (
    'cond-trip-duration-1',
    'rule-trip-duration',
    'feat-trip-duration',
    '==',
    'true'
)
ON DUPLICATE KEY UPDATE value=value;

-- 插入决策：通过
INSERT INTO decisions (id, rule_id, type, reason, action)
VALUES (
    'dec-trip-duration-1',
    'rule-trip-duration',
    'approval',
    '差旅天数合理',
    'approve'
)
ON DUPLICATE KEY UPDATE reason=reason;

-- 插入决策：拒绝
INSERT INTO decisions (id, rule_id, type, reason, action)
VALUES (
    'dec-trip-duration-2',
    'rule-trip-duration',
    'rejection',
    '差旅天数超出限制',
    'reject'
)
ON DUPLICATE KEY UPDATE reason=reason;
