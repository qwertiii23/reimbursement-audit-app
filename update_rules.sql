-- 更新规则定义到数据库
-- 从 invoice_validation_rules.grl 导入

UPDATE rules SET definition = 'rule invoice_number_required "发票号码检查" salience 110 {
    when
        data.Invoice.Number == "" && result.Passed == true
    then
        result.Passed = false;
        result.Message = "发票代码和发票号码为必填字段";
        result.Severity = "high";
        result.Priority = 110;
}' WHERE rule_code = 'RULE_INVOICE_NUMBER_REQUIRED';

UPDATE rules SET definition = 'rule invoice_date_not_future "发票日期检查" salience 101 {
    when
        data.ApplyDate.Before(data.InvoiceDateValue) && result.Passed == true
    then
        result.Passed = false;
        result.Message = "发票开票日期不能晚于报销申请日期";
        result.Severity = "high";
        result.Priority = 101;
}' WHERE rule_code = 'RULE_INVOICE_DATE_NOT_FUTURE';

UPDATE rules SET definition = 'rule invoice_date_valid "发票开票日期时效性检查" salience 100 {
    when
        data.IsInvoiceDateOlderThan6Months == true && result.Passed == true
    then
        result.Passed = false;
        result.Message = "发票开票日期超过半年，无法报销";
        result.Severity = "high";
        result.Priority = 100;
}' WHERE rule_code = 'RULE_INVOICE_DATE_VALID';

UPDATE rules SET definition = 'rule invoice_amount_positive "发票金额检查" salience 99 {
    when
        data.Invoice.Amount <= 0 && result.Passed == true
    then
        result.Passed = false;
        result.Message = "发票金额必须大于0";
        result.Severity = "high";
        result.Priority = 99;
}' WHERE rule_code = 'RULE_INVOICE_AMOUNT_POSITIVE';

UPDATE rules SET definition = 'rule accommodation_limit "住宿费限额检查" salience 90 {
    when
        data.Invoice.Category == "差旅费" && data.Invoice.SubCategory == "住宿费" && data.Invoice.Amount > 90.0 && result.Passed == true
    then
        result.Passed = false;
        result.Message = "住宿费单次不得超过90元，超出部分需特殊审批";
        result.Severity = "high";
        result.Priority = 90;
}' WHERE rule_code = 'RULE_ACCOMMODATION_LIMIT';

UPDATE rules SET definition = 'rule transportation_limit "交通费限额检查" salience 89 {
    when
        data.Invoice.Category == "差旅费" && data.Invoice.SubCategory == "交通费" && data.Invoice.Amount > 500.0 && result.Passed == true
    then
        result.Passed = false;
        result.Message = "交通费单次不得超过500元，超出部分需特殊审批";
        result.Severity = "high";
        result.Priority = 89;
}' WHERE rule_code = 'RULE_TRANSPORTATION_LIMIT';

UPDATE rules SET definition = 'rule dining_limit "餐饮费限额检查" salience 88 {
    when
        data.Invoice.Category == "差旅费" && data.Invoice.SubCategory == "餐饮费" && data.Invoice.Amount > 300.0 && result.Passed == true
    then
        result.Passed = false;
        result.Message = "餐饮费单次不得超过300元，超出部分需特殊审批";
        result.Severity = "high";
        result.Priority = 88;
}' WHERE rule_code = 'RULE_DINING_LIMIT';

UPDATE rules SET definition = 'rule office_supplies_limit "办公用品限额检查" salience 87 {
    when
        data.Invoice.Category == "办公费" && data.Invoice.SubCategory == "办公用品" && data.Invoice.Amount > 1000.0 && result.Passed == true
    then
        result.Passed = false;
        result.Message = "办公用品单次不得超过1000元，超出部分需特殊审批";
        result.Severity = "high";
        result.Priority = 87;
}' WHERE rule_code = 'RULE_OFFICE_SUPPLIES_LIMIT';

UPDATE rules SET definition = 'rule entertainment_limit "招待费限额检查" salience 86 {
    when
        data.Invoice.Category == "招待费" && data.Invoice.Amount > 2000.0 && result.Passed == true
    then
        result.Passed = false;
        result.Message = "招待费单次不得超过2000元，超出部分需特殊审批";
        result.Severity = "high";
        result.Priority = 86;
}' WHERE rule_code = 'RULE_ENTERTAINMENT_LIMIT';

UPDATE rules SET definition = 'rule large_amount_approval "大额发票审批检查" salience 80 {
    when
        data.Invoice.Amount > 5000.0 && result.Passed == true
    then
        result.Passed = false;
        result.Message = "单张发票金额超过5000元，需要特殊审批";
        result.Severity = "high";
        result.Priority = 80;
}' WHERE rule_code = 'RULE_LARGE_AMOUNT_APPROVAL';

UPDATE rules SET definition = 'rule very_large_amount_approval "超大额发票审批检查" salience 79 {
    when
        data.Invoice.Amount > 10000.0 && result.Passed == true
    then
        result.Passed = false;
        result.Message = "单张发票金额超过10000元，需要更高级别审批";
        result.Severity = "high";
        result.Priority = 79;
}' WHERE rule_code = 'RULE_VERY_LARGE_AMOUNT_APPROVAL';

UPDATE rules SET definition = 'rule tax_amount_reasonable "税额合理性检查" salience 70 {
    when
        data.Invoice.TaxAmount > data.Invoice.Amount * 0.2 && result.Passed == true
    then
        result.Passed = false;
        result.Message = "税额不应超过发票金额的20%";
        result.Severity = "medium";
        result.Priority = 70;
}' WHERE rule_code = 'RULE_TAX_AMOUNT_REASONABLE';

UPDATE rules SET definition = 'rule invoice_type_valid "发票类型检查" salience 60 {
    when
        data.Invoice.Type != "增值税专用发票" && data.Invoice.Type != "增值税普通发票" && data.Invoice.Type != "电子发票" && data.Invoice.Type != "定额发票" && result.Passed == true && data.Invoice.Type != "全电发票"
    then
        result.Passed = false;
        result.Message = "发票类型无效，支持的类型：增值税专用发票、增值税普通发票、电子发票、定额发票";
        result.Severity = "medium";
        result.Priority = 60;
}' WHERE rule_code = 'RULE_INVOICE_TYPE_VALID';

UPDATE rules SET definition = 'rule invoice_category_valid "发票类别检查" salience 59 {
    when
        data.Invoice.Category != "差旅费" && data.Invoice.Category != "办公费" && data.Invoice.Category != "招待费" && data.Invoice.Category != "培训费" && data.Invoice.Category != "采购费" && result.Passed == true
    then
        result.Passed = false;
        result.Message = "发票类别无效，支持的类别：差旅费、办公费、招待费、培训费、采购费";
        result.Severity = "medium";
        result.Priority = 59;
}' WHERE rule_code = 'RULE_INVOICE_CATEGORY_VALID';

UPDATE rules SET definition = 'rule amount_precision "金额精度检查" salience 50 {
    when
        (data.Invoice.Amount * 100) != float64(int(data.Invoice.Amount * 100)) && result.Passed == true
    then
        result.Passed = false;
        result.Message = "发票金额最多保留2位小数";
        result.Severity = "low";
        result.Priority = 50;
}' WHERE rule_code = 'RULE_AMOUNT_PRECISION';

UPDATE rules SET definition = 'rule weekend_invoice_warning "周末发票检查" salience 40 {
    when
        data.Invoice.Date != nil && (data.Invoice.Date.Weekday() == 0 || data.Invoice.Date.Weekday() == 6) && result.Passed == true
    then
        result.Passed = false;
        result.Message = "发票开票日期为周末，需要说明原因";
        result.Severity = "low";
        result.Priority = 40;
}' WHERE rule_code = 'RULE_WEEKEND_INVOICE_WARNING';

UPDATE rules SET definition = 'rule high_amount_warning "大额发票预警" salience 30 {
    when
        data.Invoice.Amount > 2000.0 && data.Invoice.Amount <= 5000.0 && result.Passed == true
    then
        result.Passed = false;
        result.Message = "单张发票金额较大，请确认合理性";
        result.Severity = "low";
        result.Priority = 30;
}' WHERE rule_code = 'RULE_HIGH_AMOUNT_WARNING';

UPDATE rules SET definition = 'rule total_amount_warning "总金额检查" salience 25 {
    when
        data.Invoice.Amount > 10000.0 && result.Passed == true
    then
        result.Passed = false;
        result.Message = "同一报销单总金额较大，请确认合理性";
        result.Severity = "low";
        result.Priority = 25;
}' WHERE rule_code = 'RULE_TOTAL_AMOUNT_WARNING';

UPDATE rules SET definition = 'rule many_invoices_warning "发票数量检查" salience 20 {
    when
        len(data.CompanyNames) > 10 && result.Passed == true
    then
        result.Passed = false;
        result.Message = "同一报销单发票数量较多，请确认合理性";
        result.Severity = "low";
        result.Priority = 20;
}' WHERE rule_code = 'RULE_MANY_INVOICES_WARNING';

UPDATE rules SET definition = 'rule remark_empty_warning "备注检查" salience 10 {
    when
        data.Invoice.Remark == "" && result.Passed == true
    then
        result.Passed = false;
        result.Message = "建议填写发票备注信息";
        result.Severity = "low";
        result.Priority = 10;
}' WHERE rule_code = 'RULE_REMARK_EMPTY_WARNING';
