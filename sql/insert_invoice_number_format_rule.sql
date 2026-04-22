-- 插入发票号码格式校验特征
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
    'feat-invoice-number-format',
    '发票号码格式校验',
    'invoice_number_format',
    '校验发票号码格式是否为8-10位纯数字',
    'boolean',
    'single',
    'invoice',
    1,
    'invoice_number_format',
    '{"min_length":8,"max_length":10,"allow_leading_zero":false}'
);

-- 插入发票号码格式校验规则
INSERT INTO rule_engine_rules (
    id,
    name,
    description,
    priority,
    enabled
) VALUES (
    'rule-invoice-number-format',
    '发票号码格式校验规则',
    '检测报销单中的发票号码格式是否为8-10位纯数字，不符合要求直接驳回',
    85,
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
    'cond-invoice-number-format',
    'rule-invoice-number-format',
    'feat-invoice-number-format',
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
    'decision-invoice-number-format',
    'rule-invoice-number-format',
    'reject',
    '发票号码格式不符合要求，必须为8-10位纯数字'
);