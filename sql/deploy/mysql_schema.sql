-- =====================================================
-- 智能报销审核系统 - MySQL 完整建表脚本
-- 适用于 MySQL 5.7+ / MySQL 8.0+
-- =====================================================

SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 1. 用户表 (users)
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` VARCHAR(36) NOT NULL COMMENT '用户ID',
  `username` VARCHAR(50) NOT NULL COMMENT '用户名',
  `password` VARCHAR(255) NOT NULL COMMENT '密码(加密存储)',
  `email` VARCHAR(100) COMMENT '邮箱',
  `real_name` VARCHAR(50) COMMENT '真实姓名',
  `role` VARCHAR(20) NOT NULL DEFAULT 'user' COMMENT '角色: admin/finance/user',
  `status` VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态: active/inactive/locked',
  `last_login` DATETIME COMMENT '最后登录时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ----------------------------
-- 2. 报销单主表 (reimbursements)
-- ----------------------------
DROP TABLE IF EXISTS `reimbursements`;
CREATE TABLE `reimbursements` (
  `id` VARCHAR(36) NOT NULL COMMENT '报销单ID',
  `user_id` VARCHAR(36) NOT NULL COMMENT '用户ID',
  `user_name` VARCHAR(100) NOT NULL COMMENT '用户姓名',
  `department` VARCHAR(100) COMMENT '所属部门',
  `applicant_level` VARCHAR(20) COMMENT '申请人级别(高管/经理/员工)',
  `type` VARCHAR(50) COMMENT '报销类型(交通/住宿/餐饮等)',
  `title` VARCHAR(200) NOT NULL COMMENT '报销标题',
  `description` TEXT COMMENT '报销描述',
  `total_amount` DECIMAL(10,2) NOT NULL COMMENT '总金额',
  `currency` VARCHAR(10) DEFAULT 'CNY' COMMENT '币种',
  `apply_date` DATE NOT NULL COMMENT '申请日期',
  `expense_date` DATE COMMENT '费用发生日期',
  `start_date` DATE COMMENT '出差开始日期',
  `end_date` DATE COMMENT '出差结束日期',
  `destination` VARCHAR(100) COMMENT '出差目的地',
  `city` VARCHAR(50) COMMENT '出差城市',
  `province` VARCHAR(50) COMMENT '出差省份',
  `travel_reason` VARCHAR(200) COMMENT '出差事由',
  `transportation` VARCHAR(50) COMMENT '交通工具',
  `project_code` VARCHAR(50) COMMENT '项目编码',
  `budget_code` VARCHAR(50) COMMENT '预算科目',
  `approval_required` TINYINT(1) DEFAULT 0 COMMENT '是否需要审批',
  `approved_by` VARCHAR(36) COMMENT '审批人ID',
  `approved_at` DATETIME COMMENT '审批时间',
  `audit_id` VARCHAR(36) COMMENT '审核ID',
  `status` VARCHAR(20) NOT NULL DEFAULT 'pending_submission' COMMENT '状态',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_audit_id` (`audit_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='报销单主表';

-- ----------------------------
-- 3. 发票表 (invoices)
-- ----------------------------
DROP TABLE IF EXISTS `invoices`;
CREATE TABLE `invoices` (
  `id` VARCHAR(36) NOT NULL COMMENT '发票ID',
  `reimbursement_id` VARCHAR(36) NOT NULL COMMENT '报销单ID',
  `type` VARCHAR(50) COMMENT '发票类型',
  `code` VARCHAR(50) COMMENT '发票代码',
  `number` VARCHAR(50) COMMENT '发票号码',
  `date` DATE COMMENT '开票日期',
  `amount` DECIMAL(10,2) NOT NULL COMMENT '发票金额',
  `tax_amount` DECIMAL(10,2) COMMENT '税额',
  `payer` VARCHAR(100) COMMENT '付款方',
  `payee` VARCHAR(100) COMMENT '收款方',
  `buyer_name` VARCHAR(100) COMMENT '购买方名称',
  `buyer_tax_no` VARCHAR(50) COMMENT '购买方税号',
  `seller_name` VARCHAR(100) COMMENT '销售方名称',
  `seller_tax_no` VARCHAR(50) COMMENT '销售方税号',
  `commodity_name` VARCHAR(200) COMMENT '商品名称',
  `specification` VARCHAR(100) COMMENT '规格型号',
  `unit` VARCHAR(20) COMMENT '单位',
  `quantity` DECIMAL(10,2) COMMENT '数量',
  `price` DECIMAL(10,2) COMMENT '单价',
  `image_path` VARCHAR(500) COMMENT '发票图片路径',
  `ocr_result` TEXT COMMENT 'OCR识别结果',
  `items` JSON COMMENT '商品明细(JSON)',
  `total_items` INT DEFAULT 0 COMMENT '商品总数',
  `main_commodity` VARCHAR(200) COMMENT '主要商品名称',
  `status` VARCHAR(20) NOT NULL DEFAULT 'pending_recognition' COMMENT '状态',
  `category` VARCHAR(50) COMMENT '发票类别',
  `sub_category` VARCHAR(50) COMMENT '发票子类别',
  `expense_type` VARCHAR(50) COMMENT '费用类型',
  `payment_method` VARCHAR(50) COMMENT '支付方式',
  `merchant_type` VARCHAR(50) COMMENT '商户类型',
  `merchant_code` VARCHAR(50) COMMENT '商户编码',
  `location` VARCHAR(100) COMMENT '消费地点',
  `city` VARCHAR(50) COMMENT '消费城市',
  `province` VARCHAR(50) COMMENT '消费省份',
  `country` VARCHAR(50) DEFAULT 'China' COMMENT '消费国家',
  `purpose` VARCHAR(200) COMMENT '消费目的',
  `project_code` VARCHAR(50) COMMENT '项目编码',
  `department_code` VARCHAR(50) COMMENT '部门编码',
  `cost_center` VARCHAR(50) COMMENT '成本中心',
  `contract_number` VARCHAR(50) COMMENT '合同编号',
  `approval_level` VARCHAR(20) COMMENT '审批级别',
  `is_reimbursable` TINYINT(1) DEFAULT 1 COMMENT '是否可报销',
  `is_personal` TINYINT(1) DEFAULT 0 COMMENT '是否个人消费',
  `is_vat` TINYINT(1) DEFAULT 0 COMMENT '是否增值税发票',
  `vat_rate` DECIMAL(5,2) DEFAULT 0.00 COMMENT '增值税率',
  `exchange_rate` DECIMAL(10,4) DEFAULT 1.0000 COMMENT '汇率',
  `original_amount` DECIMAL(10,2) DEFAULT 0.00 COMMENT '原币金额',
  `original_currency` VARCHAR(10) COMMENT '原币种',
  `receipt_number` VARCHAR(50) COMMENT '收据编号',
  `invoice_series` VARCHAR(50) COMMENT '发票系列',
  `batch_number` VARCHAR(50) COMMENT '批次号',
  `valid_from` DATE COMMENT '有效期开始',
  `valid_to` DATE COMMENT '有效期结束',
  `is_electronic` TINYINT(1) DEFAULT 0 COMMENT '是否电子发票',
  `is_duplicate` TINYINT(1) DEFAULT 0 COMMENT '是否重复发票',
  `related_invoice_id` VARCHAR(36) COMMENT '关联发票ID',
  `verification_status` VARCHAR(20) DEFAULT 'unverified' COMMENT '验证状态',
  `verification_time` DATETIME COMMENT '验证时间',
  `remarks` TEXT COMMENT '备注',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_reimbursement_id` (`reimbursement_id`),
  KEY `idx_number` (`number`),
  KEY `idx_status` (`status`),
  CONSTRAINT `fk_invoice_reimbursement` FOREIGN KEY (`reimbursement_id`) REFERENCES `reimbursements` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='发票表';

-- ----------------------------
-- 4. 审核结果表 (audit_results)
-- ----------------------------
DROP TABLE IF EXISTS `audit_results`;
CREATE TABLE `audit_results` (
  `id` VARCHAR(36) NOT NULL COMMENT '审核ID',
  `reimbursement_id` VARCHAR(36) NOT NULL COMMENT '报销单ID',
  `status` VARCHAR(20) NOT NULL COMMENT '审核状态',
  `workflow_status` VARCHAR(50) COMMENT '工作流状态',
  `rule_pass` TINYINT(1) DEFAULT 0 COMMENT '规则是否通过',
  `rag_pass` TINYINT(1) DEFAULT 0 COMMENT 'RAG是否通过',
  `final_pass` TINYINT(1) DEFAULT 0 COMMENT '最终是否通过',
  `risk_level` VARCHAR(20) COMMENT '风险等级',
  `risk_score` DECIMAL(5,4) DEFAULT 0 COMMENT '风险分数',
  `reason` TEXT COMMENT '审核原因',
  `rule_results` TEXT COMMENT '规则校验结果(JSON)',
  `rag_results` TEXT COMMENT 'RAG审核结果(JSON)',
  `suggestions` TEXT COMMENT '建议',
  `started_at` DATETIME COMMENT '开始时间',
  `completed_at` DATETIME COMMENT '完成时间',
  `duration` BIGINT COMMENT '耗时(毫秒)',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_reimbursement_id` (`reimbursement_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审核结果表';

-- ----------------------------
-- 5. 审核流程日志表 (audit_flow_logs)
-- ----------------------------
DROP TABLE IF EXISTS `audit_flow_logs`;
CREATE TABLE `audit_flow_logs` (
  `id` VARCHAR(36) NOT NULL COMMENT '日志ID',
  `reimbursement_id` VARCHAR(36) NOT NULL COMMENT '报销单ID',
  `audit_id` VARCHAR(36) NOT NULL COMMENT '审核ID',
  `flow_status` VARCHAR(50) COMMENT '流程状态',
  `flow_type` VARCHAR(20) COMMENT '流程类型',
  `operator_id` VARCHAR(36) COMMENT '操作人ID',
  `operator_name` VARCHAR(100) COMMENT '操作人姓名',
  `action` VARCHAR(50) COMMENT '操作动作',
  `reason` TEXT COMMENT '原因',
  `result` TEXT COMMENT '结果',
  `ip_address` VARCHAR(50) COMMENT 'IP地址',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_reimbursement_id` (`reimbursement_id`),
  KEY `idx_audit_id` (`audit_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审核流程日志表';

-- ----------------------------
-- 6. 规则引擎-规则表 (rule_engine_rules)
-- ----------------------------
DROP TABLE IF EXISTS `rule_engine_rules`;
CREATE TABLE `rule_engine_rules` (
  `id` VARCHAR(36) NOT NULL COMMENT '规则ID',
  `name` VARCHAR(100) NOT NULL COMMENT '规则名称',
  `description` TEXT COMMENT '规则描述',
  `priority` INT DEFAULT 0 COMMENT '优先级',
  `enabled` TINYINT(1) DEFAULT 1 COMMENT '是否启用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='规则表';

-- ----------------------------
-- 7. 规则引擎-条件表 (conditions)
-- ----------------------------
DROP TABLE IF EXISTS `conditions`;
CREATE TABLE `conditions` (
  `id` VARCHAR(36) NOT NULL COMMENT '条件ID',
  `rule_id` VARCHAR(36) NOT NULL COMMENT '规则ID',
  `feature_id` VARCHAR(36) NOT NULL COMMENT '特征ID',
  `operator` VARCHAR(20) NOT NULL COMMENT '运算符',
  `value` TEXT COMMENT '比较值',
  `logic_op` VARCHAR(10) DEFAULT 'and' COMMENT '逻辑运算符',
  `sort_order` INT DEFAULT 0 COMMENT '排序',
  PRIMARY KEY (`id`),
  KEY `idx_rule_id` (`rule_id`),
  KEY `idx_feature_id` (`feature_id`),
  CONSTRAINT `fk_condition_rule` FOREIGN KEY (`rule_id`) REFERENCES `rule_engine_rules` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_condition_feature` FOREIGN KEY (`feature_id`) REFERENCES `features` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='规则条件表';

-- ----------------------------
-- 8. 规则引擎-决策表 (decisions)
-- ----------------------------
DROP TABLE IF EXISTS `decisions`;
CREATE TABLE `decisions` (
  `id` VARCHAR(36) NOT NULL COMMENT '决策ID',
  `rule_id` VARCHAR(36) NOT NULL COMMENT '规则ID',
  `type` VARCHAR(20) NOT NULL COMMENT '决策类型',
  `reason` TEXT COMMENT '决策原因',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_rule_id` (`rule_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='规则决策表';

-- ----------------------------
-- 9. 规则引擎-特征表 (features)
-- ----------------------------
DROP TABLE IF EXISTS `features`;
CREATE TABLE `features` (
  `id` VARCHAR(36) NOT NULL COMMENT '特征ID',
  `name` VARCHAR(100) NOT NULL COMMENT '特征名称',
  `code` VARCHAR(50) NOT NULL COMMENT '特征编码',
  `description` TEXT COMMENT '特征描述',
  `type` VARCHAR(20) NOT NULL COMMENT '类型(string/number/boolean/date)',
  `value_type` VARCHAR(20) NOT NULL COMMENT '值类型(single/list)',
  `category` VARCHAR(50) COMMENT '分类',
  `enabled` TINYINT(1) DEFAULT 1 COMMENT '是否启用',
  `function_name` VARCHAR(100) COMMENT '特征函数名称',
  `function_config` JSON COMMENT '函数配置',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_category` (`category`),
  KEY `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='特征表';

-- ----------------------------
-- 10. 规则引擎-特征值表 (feature_values)
-- ----------------------------
DROP TABLE IF EXISTS `feature_values`;
CREATE TABLE `feature_values` (
  `id` VARCHAR(36) NOT NULL COMMENT '特征值ID',
  `feature_id` VARCHAR(36) NOT NULL COMMENT '特征ID',
  `value` VARCHAR(255) NOT NULL COMMENT '值',
  `label` VARCHAR(255) NOT NULL COMMENT '标签',
  `sort_order` INT DEFAULT 0 COMMENT '排序',
  `enabled` TINYINT(1) DEFAULT 1 COMMENT '是否启用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_feature_id` (`feature_id`),
  CONSTRAINT `fk_feature_value_feature` FOREIGN KEY (`feature_id`) REFERENCES `features` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='特征值表';

-- ----------------------------
-- 11. 知识库文件表 (knowledge_files)
-- ----------------------------
DROP TABLE IF EXISTS `knowledge_files`;
CREATE TABLE `knowledge_files` (
  `id` VARCHAR(36) NOT NULL COMMENT '文件ID',
  `file_name` VARCHAR(255) NOT NULL COMMENT '文件名',
  `file_path` VARCHAR(500) NOT NULL COMMENT '文件路径',
  `file_type` VARCHAR(50) NOT NULL COMMENT '文件类型',
  `file_size` BIGINT COMMENT '文件大小',
  `category` VARCHAR(100) COMMENT '分类',
  `description` TEXT COMMENT '描述',
  `uploaded_by` VARCHAR(36) COMMENT '上传人ID',
  `uploader_name` VARCHAR(100) COMMENT '上传人姓名',
  `download_count` INT DEFAULT 0 COMMENT '下载次数',
  `status` VARCHAR(20) DEFAULT 'active' COMMENT '状态',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_category` (`category`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='知识库文件表';

-- ----------------------------
-- 12. 城市级别配置表 (city_tiers)
-- ----------------------------
DROP TABLE IF EXISTS `city_tiers`;
CREATE TABLE `city_tiers` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `city_name` VARCHAR(64) NOT NULL COMMENT '城市名称',
  `city_level` VARCHAR(16) NOT NULL COMMENT '城市级别',
  `remark` VARCHAR(255) DEFAULT '' COMMENT '备注',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_city_name` (`city_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='城市级别配置表';

-- ----------------------------
-- 13. 住宿费标准表 (accommodation_standards)
-- ----------------------------
DROP TABLE IF EXISTS `accommodation_standards`;
CREATE TABLE `accommodation_standards` (
  `id` VARCHAR(36) NOT NULL COMMENT 'ID',
  `city_level` VARCHAR(16) NOT NULL COMMENT '城市级别',
  `star_rating` VARCHAR(20) NOT NULL COMMENT '星级',
  `standard_amount` DECIMAL(10,2) NOT NULL COMMENT '标准金额',
  `effective_date` DATE NOT NULL COMMENT '生效日期',
  `expiry_date` DATE COMMENT '失效日期',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_city_level` (`city_level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='住宿费标准表';

-- ----------------------------
-- 14. 餐饮费标准表 (meal_standards)
-- ----------------------------
DROP TABLE IF EXISTS `meal_standards`;
CREATE TABLE `meal_standards` (
  `id` VARCHAR(36) NOT NULL COMMENT 'ID',
  `city_level` VARCHAR(16) NOT NULL COMMENT '城市级别',
  `meal_type` VARCHAR(20) NOT NULL COMMENT '餐费类型(早餐/午餐/晚餐)',
  `standard_amount` DECIMAL(10,2) NOT NULL COMMENT '标准金额',
  `effective_date` DATE NOT NULL COMMENT '生效日期',
  `expiry_date` DATE COMMENT '失效日期',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_city_level` (`city_level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='餐饮费标准表';

-- ----------------------------
-- 15. 招待费标准表 (entertainment_standards)
-- ----------------------------
DROP TABLE IF EXISTS `entertainment_standards`;
CREATE TABLE `entertainment_standards` (
  `id` VARCHAR(36) NOT NULL COMMENT 'ID',
  `city_level` VARCHAR(16) NOT NULL COMMENT '城市级别',
  `entertainment_type` VARCHAR(20) NOT NULL COMMENT '招待类型',
  `standard_amount` DECIMAL(10,2) NOT NULL COMMENT '标准金额',
  `effective_date` DATE NOT NULL COMMENT '生效日期',
  `expiry_date` DATE COMMENT '失效日期',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_city_level` (`city_level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='招待费标准表';

-- ----------------------------
-- 16. 加班费标准表 (overtime_standards)
-- ----------------------------
DROP TABLE IF EXISTS `overtime_standards`;
CREATE TABLE `overtime_standards` (
  `id` VARCHAR(36) NOT NULL COMMENT 'ID',
  `meal_standard` DECIMAL(10,2) NOT NULL COMMENT '餐费标准',
  `transport_allowance` DECIMAL(10,2) NOT NULL COMMENT '交通补贴',
  `effective_date` DATE NOT NULL COMMENT '生效日期',
  `expiry_date` DATE COMMENT '失效日期',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='加班费标准表';

-- ----------------------------
-- 17. 交通费标准表 (transportation_standards)
-- ----------------------------
DROP TABLE IF EXISTS `transportation_standards`;
CREATE TABLE `transportation_standards` (
  `id` VARCHAR(36) NOT NULL COMMENT 'ID',
  `transport_type` VARCHAR(50) NOT NULL COMMENT '交通类型',
  `city_level` VARCHAR(16) COMMENT '城市级别',
  `distance_range` VARCHAR(50) COMMENT '距离范围',
  `standard_amount` DECIMAL(10,2) NOT NULL COMMENT '标准金额',
  `effective_date` DATE NOT NULL COMMENT '生效日期',
  `expiry_date` DATE COMMENT '失效日期',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_transport_type` (`transport_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交通费标准表';

-- ----------------------------
-- 18. 数据库迁移记录表 (migrations)
-- ----------------------------
DROP TABLE IF EXISTS `migrations`;
CREATE TABLE `migrations` (
  `id` VARCHAR(36) NOT NULL COMMENT '迁移ID',
  `name` VARCHAR(255) NOT NULL COMMENT '迁移名称',
  `applied_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '应用时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据库迁移记录表';

SET FOREIGN_KEY_CHECKS = 1;

-- =====================================================
-- 初始化基础数据
-- =====================================================

-- 插入管理员用户 (密码: admin123)
INSERT INTO `users` (`id`, `username`, `password`, `email`, `real_name`, `role`, `status`) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin@example.com', '系统管理员', 'admin', 'active');

-- 插入财务用户 (密码: finance123)
INSERT INTO `users` (`id`, `username`, `password`, `email`, `real_name`, `role`, `status`) VALUES
('550e8400-e29b-41d4-a716-446655440002', 'finance', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'finance@example.com', '财务人员', 'finance', 'active');

-- 插入测试用户 (密码: user123)
INSERT INTO `users` (`id`, `username`, `password`, `email`, `real_name`, `role`, `status`) VALUES
('550e8400-e29b-41d4-a716-446655440003', 'user', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'user@example.com', '测试用户', 'user', 'active');

-- 插入城市级别配置
INSERT INTO `city_tiers` (`city_name`, `city_level`, `remark`) VALUES
('北京', '一线城市', '首都'),
('上海', '一线城市', '直辖市'),
('广州', '一线城市', '省会'),
('深圳', '一线城市', '经济特区'),
('杭州', '新一线城市', ''),
('成都', '新一线城市', ''),
('武汉', '新一线城市', ''),
('西安', '新一线城市', ''),
('南京', '新一线城市', ''),
('重庆', '新一线城市', '直辖市'),
('天津', '二线城市', '直辖市'),
('苏州', '二线城市', ''),
('郑州', '二线城市', ''),
('长沙', '二线城市', ''),
('青岛', '二线城市', '');

-- =====================================================
-- 初始化特征数据
-- =====================================================
INSERT INTO `features` (`id`, `name`, `code`, `description`, `type`, `value_type`, `category`, `enabled`, `function_name`, `function_config`) VALUES
('detect-photoshop-feature', '是否P图', 'is_photoshopped', '检测发票图片是否经过P图处理', 'boolean', 'single', 'image', 1, 'detect_photoshop', '{}'),
('3433273d-626f-4374-8819-b661620b2f4a', '报销单金额', 'reimbursement_amount', '报销单总金额', 'number', 'single', 'amount', 1, 'reimbursement_total_amount', '{}'),
('3cb0c294-3da9-11f1-bd57-cf7f11d80696', '报销总金额', 'reimbursement_total_amount', '报销单总金额', 'number', 'single', 'amount', 1, 'reimbursement_total_amount', '{}'),
('626addbc-3da9-11f1-bd57-cf7f11d80696', '发票距今天数', 'invoice_days_from_today', '发票日期距今天的天数', 'number', 'single', 'time', 1, 'invoice_days_from_today', '{}'),
('626b0936-3da9-11f1-bd57-cf7f11d80696', '出差天数', 'trip_duration', '出差起止日期之间的天数', 'number', 'single', 'time', 1, 'trip_duration', '{}'),
('626b1de0-3da9-11f1-bd57-cf7f11d80696', '发票类型', 'invoice_type', '发票的类型', 'string', 'single', 'invoice', 1, 'invoice_type', '{}'),
('626b31fe-3da9-11f1-bd57-cf7f11d80696', '商品名称', 'commodity_name', '发票商品名称', 'string', 'single', 'invoice', 1, 'commodity_name', '{}'),
('626b3cb2-3da9-11f1-bd57-cf7f11d80696', '报销类型', 'reimbursement_type', '报销单类型', 'string', 'single', 'reimbursement', 1, 'reimbursement_type', '{}'),
('626b446e-3da9-11f1-bd57-cf7f11d80696', '发票单价', 'invoice_price', '发票商品单价', 'number', 'single', 'amount', 1, 'invoice_price', '{}'),
('4f99ee85-8817-4989-b162-4496eb15bc64', '发票分类', 'invoice_category', '发票分类', 'string', 'single', 'invoice', 1, NULL, '{}'),
('7184644e-7be8-46d8-8a2c-54769982d0b0', '发票子分类', 'invoice_subcategory', '发票子分类', 'string', 'single', 'invoice', 1, NULL, '{}'),
('a63c79a1-902d-4c94-b8fa-d69784bb4761', '发票金额', 'invoice_amount', '发票金额', 'number', 'single', 'amount', 1, 'invoice_amount', '{}'),
('feat-invoice-fraud-detection', '发票舞弊检测', 'invoice_fraud_detection', '检测发票是否存在舞弊风险', 'boolean', 'single', 'fraud', 1, 'invoice_fraud_detection', '{}'),
('feat-invoice-code-length', '发票代码长度', 'invoice_code_length', '检测发票代码长度是否合规', 'boolean', 'single', 'format', 1, 'invoice_code_length', '{}'),
('feat-invoice-type-validation', '发票类型校验', 'invoice_type_validation', '校验发票类型是否符合报销要求', 'boolean', 'single', 'validation', 1, 'invoice_type_validation', '{}'),
('feat-invoice-amount-range', '发票金额范围', 'invoice_amount_range', '校验发票金额是否在合理范围内', 'boolean', 'single', 'validation', 1, 'invoice_amount_range', '{}'),
('feat-invoice-date-validity', '开票日期有效性', 'invoice_date_validity', '校验开票日期是否有效', 'boolean', 'single', 'validation', 1, 'invoice_date_validity', '{}'),
('feat-invoice-number-format', '发票号码格式', 'invoice_number_format', '校验发票号码格式', 'boolean', 'single', 'format', 1, 'invoice_number_format', '{}'),
('feat-product-name-compliance', '商品名称合规', 'product_name_compliance', '校验商品名称是否合规', 'boolean', 'single', 'validation', 1, 'product_name_compliance', '{}'),
('feat-invoice-duplicate-check', '发票重复性校验', 'invoice_duplicate_check', '检测发票是否重复报销', 'boolean', 'single', 'duplicate', 1, 'invoice_duplicate_check', '{}'),
('feat-smart-accommodation', '智能住宿费校验', 'smart_accommodation_validation', '根据城市级别智能校验住宿费', 'boolean', 'single', 'validation', 1, 'smart_accommodation_validation', '{}'),
('feat-transportation', '交通费校验', 'transportation_validation', '校验交通费是否超标', 'boolean', 'single', 'validation', 1, 'transportation_validation', '{}'),
('feat-meal', '餐饮费校验', 'meal_validation', '校验餐饮费是否超标', 'boolean', 'single', 'validation', 1, 'meal_validation', '{}'),
('feat-entertainment', '招待费校验', 'entertainment_validation', '校验招待费是否超标', 'boolean', 'single', 'validation', 1, 'entertainment_validation', '{}'),
('feat-trip-duration', '差旅天数校验', 'trip_duration_validation', '校验出差天数是否合理', 'boolean', 'single', 'validation', 1, 'trip_duration_validation', '{}'),
('feat-overtime', '加班费校验', 'overtime_validation', '校验加班费是否合规', 'boolean', 'single', 'validation', 1, 'overtime_validation', '{}'),
('feat-merchant-type-validation', '商户类型校验', 'merchant_type_validation', '校验商户类型是否合规', 'boolean', 'single', 'validation', 1, 'merchant_type_validation', '{}'),
('feat-invoice-code-number-validation', '发票代码号码校验', 'invoice_code_number_validation', '校验发票代码号码是否正确', 'boolean', 'single', 'format', 1, 'invoice_code_number_validation', '{}');
