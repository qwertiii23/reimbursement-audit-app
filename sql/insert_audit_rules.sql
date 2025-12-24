-- 报销规则初始化数据
-- 基于企业常见报销政策和标准

-- 差旅费报销规则

-- 不要这么做 我不是有一个create接口吗 到时候调用接口再这么写 因为rule_code需要自动生成

-- 1. 住宿费报销规则 - 一线城市
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_ACCOMMODATION_TIER1',
    '一线城市住宿费上限600元',
    'rule accommodation_limit_tier1 "一线城市住宿费上限检查" salience 10 {
    when
        data.Invoice.Category == "差旅费" && data.Invoice.SubCategory == "住宿费" && data.Invoice.City in ("北京", "上海", "广州", "深圳") && data.Invoice.Amount > 600.0
    then
        result.Passed = false;
        result.Message = "一线城市住宿费超过600元上限";
        result.Severity = "medium";
        ret.AddViolation("一线城市住宿费超过600元上限", "medium", 10);
    }',
    10,
    '差旅费',
    'enabled',
    '一线城市（北京、上海、广州、深圳）住宿费每晚不得超过600元，超出部分需特殊审批',
    'system',
    NOW(),
    NOW()
);

-- 2. 住宿费报销规则 - 二线城市
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_ACCOMMODATION_TIER2',
    '二线城市住宿费上限400元',
    'rule accommodation_limit_tier2 "二线城市住宿费上限检查" salience 10 {
    when
        data.Invoice.Category == "差旅费" && data.Invoice.SubCategory == "住宿费" && isSecondTierCity(data.Invoice.City) && data.Invoice.Amount > 400.0
    then
        result.Passed = false;
        result.Message = "二线城市住宿费超过400元上限";
        result.Severity = "medium";
        ret.AddViolation("二线城市住宿费超过400元上限", "medium", 10);
    }',
    10,
    '差旅费',
    'enabled',
    '二线城市住宿费每晚不得超过400元，超出部分需特殊审批',
    'system',
    NOW(),
    NOW()
);

-- 3. 住宿费报销规则 - 三线城市
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_ACCOMMODATION_TIER3',
    '三线城市住宿费上限300元',
    'rule accommodation_limit_tier3 "三线城市住宿费上限检查" salience 10 {
    when
        data.Invoice.Category == "差旅费" && data.Invoice.SubCategory == "住宿费" && isThirdTierCity(data.Invoice.City) && data.Invoice.Amount > 300.0
    then
        result.Passed = false;
        result.Message = "三线城市住宿费超过300元上限";
        result.Severity = "medium";
        ret.AddViolation("三线城市住宿费超过300元上限", "medium", 10);
    }',
    10,
    '差旅费',
    'enabled',
    '三线城市住宿费每晚不得超过300元，超出部分需特殊审批',
    'system',
    NOW(),
    NOW()
);

-- 4. 餐饮费报销规则 - 高级管理人员
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_MEAL_EXECUTIVE',
    '高级管理人员餐饮费上限200元/天',
    'rule meal_limit_executive "高级管理人员餐饮费上限检查" salience 9 {
    when
        data.Invoice.Category == "差旅费" && data.Invoice.SubCategory == "餐饮费" && data.Reimbursement.ApplicantLevel == "高管" && data.Invoice.Amount > 200.0
    then
        result.Passed = false;
        result.Message = "高级管理人员餐饮费超过200元/天上限";
        result.Severity = "medium";
        ret.AddViolation("高级管理人员餐饮费超过200元/天上限", "medium", 9);
    }',
    9,
    '差旅费',
    'enabled',
    '高级管理人员出差期间餐饮费每天不得超过200元',
    'system',
    NOW(),
    NOW()
);

-- 5. 餐饮费报销规则 - 普通员工
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_MEAL_STAFF',
    '普通员工餐饮费上限100元/天',
    'rule meal_limit_staff "普通员工餐饮费上限检查" salience 9 {
    when
        data.Invoice.Category == "差旅费" && data.Invoice.SubCategory == "餐饮费" && data.Reimbursement.ApplicantLevel == "员工" && data.Invoice.Amount > 100.0
    then
        result.Passed = false;
        result.Message = "普通员工餐饮费超过100元/天上限";
        result.Severity = "medium";
        ret.AddViolation("普通员工餐饮费超过100元/天上限", "medium", 9);
    }',
    9,
    '差旅费',
    'enabled',
    '普通员工出差期间餐饮费每天不得超过100元',
    'system',
    NOW(),
    NOW()
);

-- 6. 交通费报销规则 - 飞机舱位限制
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_FLIGHT_CLASS',
    '飞机票仅限经济舱',
    'rule flight_class_limit "飞机舱位限制检查" salience 8 {
    when
        data.Invoice.Category == "差旅费" && data.Invoice.SubCategory == "交通费" && data.Invoice.MerchantType == "航空公司" && !isEconomyClass(data.Invoice.Description)
    then
        result.Passed = false;
        result.Message = "飞机票仅限经济舱，商务舱和头等舱需特殊审批";
        result.Severity = "high";
        ret.AddViolation("飞机票仅限经济舱，商务舱和头等舱需特殊审批", "high", 8);
    }',
    8,
    '差旅费',
    'enabled',
    '除高管外，所有员工乘坐飞机仅限经济舱，商务舱和头等舱需特殊审批',
    'system',
    NOW(),
    NOW()
);

-- 7. 交通费报销规则 - 高铁座位限制
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_TRAIN_CLASS',
    '高铁仅限二等座',
    'rule train_class_limit "高铁座位限制检查" salience 8 {
    when
        data.Invoice.Category == "差旅费" && data.Invoice.SubCategory == "交通费" && data.Invoice.MerchantType == "铁路公司" && !isSecondClass(data.Invoice.Description)
    then
        result.Passed = false;
        result.Message = "高铁仅限二等座，一等座和商务座需特殊审批";
        result.Severity = "medium";
        ret.AddViolation("高铁仅限二等座，一等座和商务座需特殊审批", "medium", 8);
    }',
    8,
    '差旅费',
    'enabled',
    '除高管外，所有员工乘坐高铁仅限二等座，一等座和商务座需特殊审批',
    'system',
    NOW(),
    NOW()
);

-- 发票基本规则

-- 8. 发票时效性规则 - 普通发票
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_INVOICE_TIMELINESS_COMMON',
    '普通发票6个月内有效',
    'rule invoice_timeliness_common "普通发票时效性检查" salience 20 {
    when
        data.Invoice.IsVAT == false && daysBetween(data.Invoice.Date, data.Reimbursement.ApplyDate) > 180
    then
        result.Passed = false;
        result.Message = "普通发票开具日期超过6个月，无法报销";
        result.Severity = "high";
        ret.AddViolation("普通发票开具日期超过6个月，无法报销", "high", 20);
    }',
    20,
    '发票校验',
    'enabled',
    '普通发票必须在开具日期后6个月内提交报销，超过期限无法报销',
    'system',
    NOW(),
    NOW()
);

-- 9. 发票时效性规则 - 增值税专用发票
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_INVOICE_TIMELINESS_VAT',
    '增值税专用发票180天内有效',
    'rule invoice_timeliness_vat "增值税专用发票时效性检查" salience 20 {
    when
        data.Invoice.IsVAT == true && daysBetween(data.Invoice.Date, data.Reimbursement.ApplyDate) > 180
    then
        result.Passed = false;
        result.Message = "增值税专用发票开具日期超过180天，无法认证抵扣";
        result.Severity = "high";
        ret.AddViolation("增值税专用发票开具日期超过180天，无法认证抵扣", "high", 20);
    }',
    20,
    '发票校验',
    'enabled',
    '增值税专用发票必须在开具日期后180天内提交报销，超过期限无法认证抵扣',
    'system',
    NOW(),
    NOW()
);

-- 10. 发票金额规则 - 整数金额检查
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_INVOICE_ROUND_AMOUNT',
    '整数金额发票需特别审核',
    'rule invoice_round_amount "整数金额发票检查" salience 5 {
    when
        data.Invoice.Amount == Math.round(data.Invoice.Amount) && data.Invoice.Amount > 1000.0
    then
        result.Passed = true;
        result.Message = "整数金额发票需特别审核";
        result.Severity = "low";
        ret.AddViolation("整数金额发票需特别审核", "low", 5);
    }',
    5,
    '发票校验',
    'enabled',
    '金额为整数的发票（特别是大额发票）需要特别审核，防止虚开发票',
    'system',
    NOW(),
    NOW()
);

-- 11. 发票重复报销检查
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_DUPLICATE_INVOICE',
    '发票重复报销检查',
    'rule duplicate_invoice_check "发票重复报销检查" salience 30 {
    when
        data.Invoice.IsDuplicate == true || isDuplicateInvoice(data.Invoice.Code, data.Invoice.Number)
    then
        result.Passed = false;
        result.Message = "发票已报销，不能重复提交";
        result.Severity = "high";
        ret.AddViolation("发票已报销，不能重复提交", "high", 30);
    }',
    30,
    '发票校验',
    'enabled',
    '检查发票是否已经报销过，防止重复报销',
    'system',
    NOW(),
    NOW()
);

-- 招待费报销规则

-- 12. 招待费总额限制
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_ENTERTAINMENT_MONTHLY_LIMIT',
    '招待费月度总额限制',
    'rule entertainment_monthly_limit "招待费月度总额限制检查" salience 7 {
    when
        data.Invoice.Category == "招待费" && getMonthlyEntertainmentTotal(data.Reimbursement.UserID, data.Reimbursement.ApplyDate) > 2000.0
    then
        result.Passed = false;
        result.Message = "招待费月度总额超过2000元上限";
        result.Severity = "medium";
        ret.AddViolation("招待费月度总额超过2000元上限", "medium", 7);
    }',
    7,
    '招待费',
    'enabled',
    '员工每月招待费总额不得超过2000元，超出部分需特殊审批',
    'system',
    NOW(),
    NOW()
);

-- 13. 招待费单次限额
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_ENTERTAINMENT_SINGLE_LIMIT',
    '招待费单次限额500元',
    'rule entertainment_single_limit "招待费单次限额检查" salience 7 {
    when
        data.Invoice.Category == "招待费" && data.Invoice.Amount > 500.0
    then
        result.Passed = false;
        result.Message = "招待费单次超过500元上限";
        result.Severity = "medium";
        ret.AddViolation("招待费单次超过500元上限", "medium", 7);
    }',
    7,
    '招待费',
    'enabled',
    '单次招待费不得超过500元，超出部分需特殊审批',
    'system',
    NOW(),
    NOW()
);

-- 办公费报销规则

-- 14. 办公用品单次限额
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_OFFICE_SINGLE_LIMIT',
    '办公用品单次限额1000元',
    'rule office_single_limit "办公用品单次限额检查" salience 6 {
    when
        data.Invoice.Category == "办公费" && data.Invoice.SubCategory == "办公用品" && data.Invoice.Amount > 1000.0
    then
        result.Passed = false;
        result.Message = "办公用品单次采购超过1000元上限";
        result.Severity = "medium";
        ret.AddViolation("办公用品单次采购超过1000元上限", "medium", 6);
    }',
    6,
    '办公费',
    'enabled',
    '办公用品单次采购不得超过1000元，超出部分需特殊审批',
    'system',
    NOW(),
    NOW()
);

-- 15. 办公费月度限额
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_OFFICE_MONTHLY_LIMIT',
    '办公费月度限额3000元',
    'rule office_monthly_limit "办公费月度限额检查" salience 6 {
    when
        data.Invoice.Category == "办公费" && data.Invoice.SubCategory == "办公用品" && getMonthlyOfficeTotal(data.Reimbursement.UserID, data.Reimbursement.ApplyDate) > 3000.0
    then
        result.Passed = false;
        result.Message = "办公费月度总额超过3000元上限";
        result.Severity = "medium";
        ret.AddViolation("办公费月度总额超过3000元上限", "medium", 6);
    }',
    6,
    '办公费',
    'enabled',
    '员工每月办公费总额不得超过3000元，超出部分需特殊审批',
    'system',
    NOW(),
    NOW()
);

-- 通讯费报销规则

-- 16. 通讯费月度限额
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_COMMUNICATION_MONTHLY_LIMIT',
    '通讯费月度限额200元',
    'rule communication_monthly_limit "通讯费月度限额检查" salience 6 {
    when
        data.Invoice.Category == "通讯费" && getMonthlyCommunicationTotal(data.Reimbursement.UserID, data.Reimbursement.ApplyDate) > 200.0
    then
        result.Passed = false;
        result.Message = "通讯费月度总额超过200元上限";
        result.Severity = "medium";
        ret.AddViolation("通讯费月度总额超过200元上限", "medium", 6);
    }',
    6,
    '通讯费',
    'enabled',
    '员工每月通讯费总额不得超过200元，超出部分需特殊审批',
    'system',
    NOW(),
    NOW()
);

-- 17. 周末消费规则
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_WEEKEND_CONSUMPTION',
    '周末消费需特别说明',
    'rule weekend_consumption "周末消费检查" salience 4 {
    when
        isWeekend(data.Invoice.Date) && !isBusinessTravel(data.Reimbursement.UserID, data.Invoice.Date)
    then
        result.Passed = true;
        result.Message = "周末消费需提供特别说明";
        result.Severity = "low";
        ret.AddViolation("周末消费需提供特别说明", "low", 4);
    }',
    4,
    '所有类目',
    'enabled',
    '周末发生的消费（除非出差期间）需要提供特别说明',
    'system',
    NOW(),
    NOW()
);

-- 18. 大额消费规则
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_LARGE_AMOUNT',
    '大额消费需特别审批',
    'rule large_amount "大额消费检查" salience 15 {
    when
        data.Invoice.Amount > 5000.0
    then
        result.Passed = false;
        result.Message = "单笔消费超过5000元需特别审批";
        result.Severity = "high";
        ret.AddViolation("单笔消费超过5000元需特别审批", "high", 15);
    }',
    15,
    '所有类目',
    'enabled',
    '单笔消费超过5000元需要提供特别审批',
    'system',
    NOW(),
    NOW()
);

-- 19. 连号发票规则
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_CONSECUTIVE_INVOICE',
    '连号发票需特别审核',
    'rule consecutive_invoice "连号发票检查" salience 12 {
    when
        isConsecutiveInvoice(data.InvoiceNumbers)
    then
        result.Passed = true;
        result.Message = "连号发票需特别审核";
        result.Severity = "medium";
        ret.AddViolation("连号发票需特别审核", "medium", 12);
    }',
    12,
    '发票校验',
    'enabled',
    '连号发票需要特别审核，防止拆分报销',
    'system',
    NOW(),
    NOW()
);

-- 20. 三单匹配规则
INSERT INTO audit_rules (
    id, 
    rule_code, 
    rule_name, 
    rule_content, 
    priority, 
    category, 
    status, 
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_THREE_DOCUMENT_MATCH',
    '三单匹配检查',
    'rule three_document_match "三单匹配检查" salience 25 {
    when
        data.Invoice.Amount > 1000.0 && !hasOrderAndReceipt(data.Invoice.ID)
    then
        result.Passed = false;
        result.Message = "大额消费需提供订单、发票、收据三单匹配";
        result.Severity = "high";
        ret.AddViolation("大额消费需提供订单、发票、收据三单匹配", "high", 25);
    }',
    25,
    '发票校验',
    'enabled',
    '金额超过1000元的消费需要提供订单、发票、收据三单匹配',
    'system',
    NOW(),
    NOW()
);