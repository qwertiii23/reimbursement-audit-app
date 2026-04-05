-- 插入发票重复性校验特征
INSERT INTO features (id, name, code, type, value_type, category, function_name, function_config, description)
VALUES (
    'feat-invoice-duplicate-check',
    '发票重复性校验',
    'invoice_duplicate_check',
    'boolean',
    'boolean',
    'validation',
    'invoice_duplicate_check',
    '{"check_amount": true, "check_date": false}',
    '检查发票是否已报销，防止重复报销'
)
ON DUPLICATE KEY UPDATE name=name;

-- 插入发票重复性校验规则
INSERT INTO rule_engine_rules (id, name, description, priority, enabled)
VALUES (
    'rule-invoice-duplicate-check',
    '发票重复性校验规则',
    '检查发票是否已报销，防止重复报销',
    95,
    1
)
ON DUPLICATE KEY UPDATE name=name;

-- 插入条件：发票未重复
INSERT INTO conditions (id, rule_id, feature_id, operator, value)
VALUES (
    'cond-invoice-duplicate-check-1',
    'rule-invoice-duplicate-check',
    'feat-invoice-duplicate-check',
    '==',
    'false'
)
ON DUPLICATE KEY UPDATE value=value;

-- 插入决策：通过
INSERT INTO decisions (id, rule_id, type, reason, action)
VALUES (
    'dec-invoice-duplicate-check-1',
    'rule-invoice-duplicate-check',
    'approval',
    '发票未重复',
    'approve'
)
ON DUPLICATE KEY UPDATE reason=reason;

-- 插入决策：拒绝
INSERT INTO decisions (id, rule_id, type, reason, action)
VALUES (
    'dec-invoice-duplicate-check-2',
    'rule-invoice-duplicate-check',
    'rejection',
    '发票已重复报销',
    'reject'
)
ON DUPLICATE KEY UPDATE reason=reason;
