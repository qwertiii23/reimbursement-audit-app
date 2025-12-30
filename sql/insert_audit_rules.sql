-- 报销规则初始化数据
-- 基于实际发票字段设计的实用规则

-- 1. 发票时效性规则 - 开票日期必须在半年内
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_INVOICE_DATE_VALID',
    '发票开票日期必须在半年内',
    'rule invoice_date_valid "发票开票日期时效性检查" salience 100 {
    when
        data.Invoice.Date != nil && data.ApplyDate.Sub(*data.Invoice.Date) > 4320h
    then
        result.Passed = false;
        result.Message = "发票开票日期超过半年，无法报销";
        result.Severity = "high";
        result.Priority = 100;
    }',
    100,
    '时效性校验',
    'enabled',
    '发票开票日期必须在半年（180天）内，超过期限无法报销',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 2. 发票金额必须大于0
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_INVOICE_AMOUNT_POSITIVE',
    '发票金额必须大于0',
    'rule invoice_amount_positive "发票金额检查" salience 99 {
    when
        data.Invoice.Amount <= 0
    then
        result.Passed = false;
        result.Message = "发票金额必须大于0";
        result.Severity = "high";
        result.Priority = 99;
    }',
    99,
    '金额校验',
    'enabled',
    '发票金额必须大于0',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 3. 发票号码不能为空
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_INVOICE_NUMBER_REQUIRED',
    '发票号码不能为空',
    'rule invoice_number_required "发票号码检查" salience 98 {
    when
        data.Invoice.Number == ""
    then
        result.Passed = false;
        result.Message = "发票号码不能为空";
        result.Severity = "high";
        result.Priority = 98;
    }',
    98,
    '基础校验',
    'enabled',
    '发票号码为必填字段',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 4. 发票类型不能为空
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_INVOICE_TYPE_REQUIRED',
    '发票类型不能为空',
    'rule invoice_type_required "发票类型检查" salience 97 {
    when
        data.Invoice.Type == ""
    then
        result.Passed = false;
        result.Message = "发票类型不能为空";
        result.Severity = "high";
        result.Priority = 97;
    }',
    97,
    '基础校验',
    'enabled',
    '发票类型为必填字段',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 5. 住宿费单次限额200元
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_ACCOMMODATION_LIMIT',
    '住宿费单次限额200元',
    'rule accommodation_limit "住宿费限额检查" salience 90 {
    when
        data.Invoice.Category == "差旅费" && data.Invoice.SubCategory == "住宿费" && data.Invoice.Amount > 200.0
    then
        result.Passed = false;
        result.Message = "住宿费超过200元上限";
        result.Severity = "medium";
        result.Priority = 90;
    }',
    90,
    '金额校验',
    'enabled',
    '住宿费单次不得超过200元，超出部分需特殊审批',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 6. 餐饮费单次限额100元
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_MEAL_LIMIT',
    '餐饮费单次限额100元',
    'rule meal_limit "餐饮费限额检查" salience 89 {
    when
        data.Invoice.Category == "差旅费" && data.Invoice.SubCategory == "餐饮费" && data.Invoice.Amount > 100.0
    then
        result.Passed = false;
        result.Message = "餐饮费超过100元上限";
        result.Severity = "medium";
        result.Priority = 89;
    }',
    89,
    '金额校验',
    'enabled',
    '餐饮费单次不得超过100元，超出部分需特殊审批',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 7. 交通费单次限额500元
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_TRANSPORT_LIMIT',
    '交通费单次限额500元',
    'rule transport_limit "交通费限额检查" salience 88 {
    when
        data.Invoice.Category == "差旅费" && data.Invoice.SubCategory == "交通费" && data.Invoice.Amount > 500.0
    then
        result.Passed = false;
        result.Message = "交通费超过500元上限";
        result.Severity = "medium";
        result.Priority = 88;
    }',
    88,
    '金额校验',
    'enabled',
    '交通费单次不得超过500元，超出部分需特殊审批',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 8. 招待费单次限额500元
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_ENTERTAINMENT_LIMIT',
    '招待费单次限额500元',
    'rule entertainment_limit "招待费限额检查" salience 87 {
    when
        data.Invoice.Category == "招待费" && data.Invoice.Amount > 500.0
    then
        result.Passed = false;
        result.Message = "招待费超过500元上限";
        result.Severity = "medium";
        result.Priority = 87;
    }',
    87,
    '金额校验',
    'enabled',
    '招待费单次不得超过500元，超出部分需特殊审批',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 9. 办公费单次限额1000元
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_OFFICE_LIMIT',
    '办公费单次限额1000元',
    'rule office_limit "办公费限额检查" salience 86 {
    when
        data.Invoice.Category == "办公费" && data.Invoice.Amount > 1000.0
    then
        result.Passed = false;
        result.Message = "办公费超过1000元上限";
        result.Severity = "medium";
        result.Priority = 86;
    }',
    86,
    '金额校验',
    'enabled',
    '办公费单次不得超过1000元，超出部分需特殊审批',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 10. 通讯费单次限额200元
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_COMMUNICATION_LIMIT',
    '通讯费单次限额200元',
    'rule communication_limit "通讯费限额检查" salience 85 {
    when
        data.Invoice.Category == "通讯费" && data.Invoice.Amount > 200.0
    then
        result.Passed = false;
        result.Message = "通讯费超过200元上限";
        result.Severity = "medium";
        result.Priority = 85;
    }',
    85,
    '金额校验',
    'enabled',
    '通讯费单次不得超过200元，超出部分需特殊审批',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 11. 大额消费检查 - 超过5000元需特别审批
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_LARGE_AMOUNT',
    '大额消费需特别审批',
    'rule large_amount "大额消费检查" salience 80 {
    when
        data.Invoice.Amount > 5000.0
    then
        result.Passed = false;
        result.Message = "单笔消费超过5000元需特别审批";
        result.Severity = "high";
        result.Priority = 80;
    }',
    80,
    '金额校验',
    'enabled',
    '单笔消费超过5000元需要提供特别审批',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 12. 发票税额不能超过发票金额
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_TAX_AMOUNT_VALID',
    '发票税额不能超过发票金额',
    'rule tax_amount_valid "税额检查" salience 75 {
    when
        data.Invoice.TaxAmount > data.Invoice.Amount
    then
        result.Passed = false;
        result.Message = "发票税额不能超过发票金额";
        result.Severity = "high";
        result.Priority = 75;
    }',
    75,
    '税额校验',
    'enabled',
    '发票税额必须小于等于发票金额',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 13. 发票购买方名称不能为空
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_BUYER_NAME_REQUIRED',
    '发票购买方名称不能为空',
    'rule buyer_name_required "购买方名称检查" salience 70 {
    when
        data.Invoice.BuyerName == ""
    then
        result.Passed = false;
        result.Message = "发票购买方名称不能为空";
        result.Severity = "high";
        result.Priority = 70;
    }',
    70,
    '基础校验',
    'enabled',
    '发票购买方名称为必填字段',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 14. 发票销售方名称不能为空
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_SELLER_NAME_REQUIRED',
    '发票销售方名称不能为空',
    'rule seller_name_required "销售方名称检查" salience 69 {
    when
        data.Invoice.SellerName == ""
    then
        result.Passed = false;
        result.Message = "发票销售方名称不能为空";
        result.Severity = "high";
        result.Priority = 69;
    }',
    69,
    '基础校验',
    'enabled',
    '发票销售方名称为必填字段',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 15. 发票商品名称不能为空
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_COMMODITY_NAME_REQUIRED',
    '发票商品名称不能为空',
    'rule commodity_name_required "商品名称检查" salience 68 {
    when
        data.Invoice.CommodityName == ""
    then
        result.Passed = false;
        result.Message = "发票商品名称不能为空";
        result.Severity = "medium";
        result.Priority = 68;
    }',
    68,
    '基础校验',
    'enabled',
    '发票商品名称为必填字段',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 16. 整数金额发票需特别审核（金额大于1000元且为整数）
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_ROUND_AMOUNT',
    '整数金额发票需特别审核',
    'rule round_amount "整数金额检查" salience 50 {
    when
        data.Invoice.Amount > 1000.0 && data.Invoice.Amount == float64(int(data.Invoice.Amount))
    then
        result.Passed = true;
        result.Message = "整数金额发票需特别审核";
        result.Severity = "low";
        result.Priority = 50;
    }',
    50,
    '风险提示',
    'enabled',
    '金额为整数的发票（特别是大额发票）需要特别审核，防止虚开发票',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 17. 发票图片路径不能为空
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_IMAGE_PATH_REQUIRED',
    '发票图片路径不能为空',
    'rule image_path_required "图片路径检查" salience 60 {
    when
        data.Invoice.ImagePath == ""
    then
        result.Passed = false;
        result.Message = "发票图片路径不能为空";
        result.Severity = "high";
        result.Priority = 60;
    }',
    60,
    '基础校验',
    'enabled',
    '发票图片路径为必填字段',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 18. 发票状态必须为已识别
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_STATUS_VALID',
    '发票状态必须为已识别',
    'rule status_valid "发票状态检查" salience 65 {
    when
        data.Invoice.Status != "recognized"
    then
        result.Passed = false;
        result.Message = "发票状态必须为已识别";
        result.Severity = "high";
        result.Priority = 65;
    }',
    65,
    '状态校验',
    'enabled',
    '发票必须完成OCR识别才能进行审核',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 19. 培训费单次限额2000元
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_TRAINING_LIMIT',
    '培训费单次限额2000元',
    'rule training_limit "培训费限额检查" salience 84 {
    when
        data.Invoice.Category == "培训费" && data.Invoice.Amount > 2000.0
    then
        result.Passed = false;
        result.Message = "培训费超过2000元上限";
        result.Severity = "medium";
        result.Priority = 84;
    }',
    84,
    '金额校验',
    'enabled',
    '培训费单次不得超过2000元，超出部分需特殊审批',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);

-- 20. 会议费单次限额3000元
INSERT INTO rules (
    id, 
    rule_code, 
    name, 
    definition, 
    priority, 
    category, 
    status, 
    description,
    type,
    enabled,
    created_by,
    created_at,
    updated_at
) VALUES (
    UUID(),
    'RULE_MEETING_LIMIT',
    '会议费单次限额3000元',
    'rule meeting_limit "会议费限额检查" salience 83 {
    when
        data.Invoice.Category == "会议费" && data.Invoice.Amount > 3000.0
    then
        result.Passed = false;
        result.Message = "会议费超过3000元上限";
        result.Severity = "medium";
        result.Priority = 83;
    }',
    83,
    '金额校验',
    'enabled',
    '会议费单次不得超过3000元，超出部分需特殊审批',
    'invoice_validation',
    1,
    'system',
    NOW(),
    NOW()
);
