-- 插入发票舞弊检测特征
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
    'feat-invoice-fraud-detection',
    '发票舞弊检测',
    'invoice_fraud_detection',
    '检测发票图片是否存在P图/篡改痕迹，支持多张发票检测',
    'boolean',
    'single',
    'invoice',
    1,
    'invoice_fraud_detection',
    '{"confidence_threshold":0.7,"check_all_images":true,"detection_prompt":"请仔细分析这张发票图片，判断是否存在明显的P图痕迹、篡改或伪造。请从以下几个方面分析：1. 文字是否清晰自然，字体是否一致 2. 数字和金额是否合理，有无涂改痕迹 3. 印章和签名是否真实，有无PS痕迹 4. 整体布局是否合理，有无拼接痕迹 5. 图片质量是否异常，有无模糊或失真。请给出你的判断结果（存在P图/不存在P图）和置信度（0-1之间）。"}'
);

-- 插入发票舞弊检测规则
INSERT INTO rule_engine_rules (
    id,
    name,
    description,
    priority,
    enabled
) VALUES (
    'rule-invoice-fraud-detection',
    '发票舞弊检测规则',
    '检测报销单中的发票是否存在P图、篡改或伪造痕迹，发现舞弊直接驳回',
    100,
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
    'cond-invoice-fraud-detection',
    'rule-invoice-fraud-detection',
    'feat-invoice-fraud-detection',
    'eq',
    'true',
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
    'decision-invoice-fraud-detection',
    'rule-invoice-fraud-detection',
    'reject',
    '发票存在P图/篡改痕迹，不符合报销要求'
);
