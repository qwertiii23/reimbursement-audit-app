ALTER TABLE features ADD COLUMN function_name VARCHAR(100) COMMENT '特征函数名称' AFTER enabled;
ALTER TABLE features ADD COLUMN function_config JSON COMMENT '特征函数配置' AFTER function_name;

INSERT INTO features (id, name, code, description, type, value_type, category, enabled, function_name, function_config) VALUES
(UUID(), '报销总金额', 'reimbursement_total_amount', '报销单的总金额', 'number', 'single', 'amount', TRUE, 'reimbursement_total_amount', '{}'),
(UUID(), '发票金额', 'invoice_amount', '单张发票的金额', 'number', 'single', 'amount', TRUE, 'invoice_amount', '{}'),
(UUID(), '发票距今天数', 'invoice_days_from_today', '发票日期距今天的天数', 'number', 'single', 'time', TRUE, 'invoice_days_from_today', '{}'),
(UUID(), '出差天数', 'trip_duration', '出差起止日期之间的天数', 'number', 'single', 'time', TRUE, 'trip_duration', '{}'),
(UUID(), '发票类型', 'invoice_type', '发票的类型（增值税普通发票、增值税专用发票等）', 'string', 'single', 'invoice', TRUE, 'invoice_type', '{}'),
(UUID(), '商品名称', 'commodity_name', '发票商品名称', 'string', 'single', 'invoice', TRUE, 'commodity_name', '{}'),
(UUID(), '商户类型', 'merchant_type', '发票商户类型', 'string', 'single', 'invoice', TRUE, 'merchant_type', '{}'),
(UUID(), '报销类型', 'reimbursement_type', '报销单类型（差旅费、业务招待费等）', 'string', 'single', 'reimbursement', TRUE, 'reimbursement_type', '{}'),
(UUID(), '申请人级别', 'applicant_level', '申请人的职级', 'string', 'single', 'user', TRUE, 'applicant_level', '{}'),
(UUID(), '开票日期有效性', 'invoice_date_validity', '校验开票日期是否在有效范围内', 'boolean', 'single', 'time', TRUE, 'invoice_date_validity', '{"max_days_ago": 365}'),
(UUID(), '发票金额范围', 'invoice_amount_range', '校验发票金额是否在合理范围内', 'boolean', 'single', 'amount', TRUE, 'invoice_amount_range', '{"min_amount": 0, "max_amount": 100000}'),
(UUID(), '发票单价', 'invoice_price', '发票商品单价', 'number', 'single', 'amount', TRUE, 'invoice_price', '{}');
