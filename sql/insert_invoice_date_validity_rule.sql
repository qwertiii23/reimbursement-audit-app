-- 插入开票日期有效性校验特征
INSERT INTO features (
    id,
    name,
    code,
    description,
    type,
    value_type,
    category,
    enabled,
    function_name,
    function_config
) VALUES (
    'feat-invoice-date-validity',
    '开票日期有效性校验',
    'invoice_date_validity',
    '校验开票日期是否在有效范围内（当前日期往前推365天至当前日期）',
    'boolean',
    'single',
    'invoice',
    1,
    'invoice_date_validity',
    '{"max_days_ago":365}'
);

-- 插入开票日期有效性校验规则
INSERT INTO rule_engine_rules (
    id,
    name,
    description,
    priority,
    enabled
) VALUES (
    'rule-invoice-date-validity',
    '开票日期有效性校验规则',
    '检测报销单中的开票日期是否在有效范围内（当前日期往前推365天至当前日期），过期则驳回',
    80,
    1
);

-- 插入规则条件
INSERT INTO conditions (
    id,
    rule_id,
    feature_id,
    operator,
    value,
    logic_op,
    sort_order
) VALUES (
    'cond-invoice-date-validity',
    'rule-invoice-date-validity',
    'feat-invoice-date-validity',
    'eq',
    'false',
    'and',
    1
);

-- 插入规则决策
INSERT INTO decisions (
    id,
    rule_id,
    type,
    reason
) VALUES (
    'decision-invoice-date-validity',
    'rule-invoice-date-validity',
    'reject',
    '开票日期不在有效范围内，必须为当前日期往前推365天至当前日期'
);