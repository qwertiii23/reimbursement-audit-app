-- 插入商品名称合规校验特征
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
    'feat-product-name-compliance',
    '商品名称合规校验',
    'product_name_compliance',
    '检测商品名称是否包含敏感/禁报关键词（烟酒、奢侈品、黄金、娱乐服务等）',
    'boolean',
    'single',
    'invoice',
    1,
    'product_name_compliance',
    '{"sensitive_keywords":["烟","酒","香烟","白酒","红酒","啤酒","奢侈品","黄金","珠宝","钻戒","娱乐服务","KTV","酒吧","夜总会","洗浴","按摩","足浴","会所"],"match_mode":"contains","case_sensitive":false}'
);

-- 插入商品名称合规校验规则
INSERT INTO rule_engine_rules (
    id,
    name,
    description,
    priority,
    enabled
) VALUES (
    'rule-product-name-compliance',
    '商品名称合规校验规则',
    '检测报销单中的商品名称是否包含敏感/禁报关键词，包含敏感关键词直接驳回',
    90,
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
    'cond-product-name-compliance',
    'rule-product-name-compliance',
    'feat-product-name-compliance',
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
    'decision-product-name-compliance',
    'rule-product-name-compliance',
    'reject',
    '商品名称包含敏感/禁报关键词，不符合报销要求'
);