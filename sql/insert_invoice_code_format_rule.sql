-- 插入发票代码格式校验规则
-- 规则名称：发票代码格式校验
-- 特征值：发票代码长度
-- 条件：包含
-- 参考值：10-12 位纯数字
-- 说明：发票代码需符合 10-12 位纯数字的格式规范，不符合则直接驳回

-- 1. 插入特征值（发票代码长度）
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
    function_config,
    created_at,
    updated_at
) VALUES (
    'feat-invoice-code-length',
    '发票代码长度',
    'invoice_code_length',
    '检查发票代码长度（10-12位纯数字）',
    'boolean',
    'single',
    'invoice',
    true,
    'invoice_code_length',
    '{"min_length": 10, "max_length": 12}',
    NOW(),
    NOW()
);

-- 2. 插入规则
INSERT INTO rule_engine_rules (
    id,
    name,
    description,
    priority,
    enabled,
    created_at,
    updated_at
) VALUES (
    'rule-invoice-code-format',
    '发票代码格式校验',
    '发票代码需符合 10-12 位纯数字的格式规范，不符合则直接驳回',
    10,
    true,
    NOW(),
    NOW()
);

-- 3. 插入条件
INSERT INTO conditions (
    id,
    rule_id,
    feature_id,
    operator,
    value,
    logic_op,
    sort_order
) VALUES (
    'cond-invoice-code-length',
    'rule-invoice-code-format',
    'feat-invoice-code-length',
    'eq',
    'false',
    'and',
    0
);

-- 4. 插入决策
INSERT INTO decisions (
    id,
    rule_id,
    type,
    reason,
    created_at,
    updated_at
) VALUES (
    'dec-invoice-code-format',
    'rule-invoice-code-format',
    'reject',
    '发票代码格式不符合规范，需为10-12位纯数字',
    NOW(),
    NOW()
);
