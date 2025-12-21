-- ------------------------------------------------------
-- 智能报销审核系统 - MySQL 5.6 适配版初始化脚本
-- 文件名：reimbursement-audit.sql
-- 适配版本：MySQL 5.6（兼容低版本特性）
-- 核心说明：
-- 1. 移除 JSON 类型，改用 TEXT 存储 JSON 格式字符串
-- 2. 移除原生 UUID() 函数，手动生成示例 UUID（实际开发需通过业务层生成）
-- 3. 保留核心业务表、外键约束、索引设计，确保功能完整性
-- 4. 向量嵌入以 BASE64 编码字符串存储（业务层需解码使用）
-- ------------------------------------------------------

-- 关闭外键约束检查（避免创建表顺序导致的外键报错）
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 1. 城市级别配置表（city_levels）
-- 作用：存储城市与级别映射关系，为规则引擎提供“城市级别”判定依据
-- ----------------------------
DROP TABLE IF EXISTS `city_levels`;
CREATE TABLE `city_levels` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `city_name` VARCHAR(64) NOT NULL COMMENT '城市名称（如：北京、上海、杭州）',
  `city_level` VARCHAR(16) NOT NULL COMMENT '城市级别（枚举值：一线城市/新一线城市/二线城市/其他城市）',
  `remark` VARCHAR(255) DEFAULT '' COMMENT '备注（如：2024年第一财经榜单认定）',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_city_name` (`city_name`) COMMENT '城市名称唯一约束，避免重复配置',
  KEY `idx_city_level` (`city_level`) COMMENT '按城市级别查询索引，优化规则引擎查询效率'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='城市级别配置表（规则引擎依赖）';

-- ----------------------------
-- 2. 报销单主表（reimbursements）
-- 作用：核心主表，存储报销单核心信息，作为所有关联数据的“主索引”
-- ----------------------------
DROP TABLE IF EXISTS `reimbursements`;
CREATE TABLE `reimbursements` (
  `id` VARCHAR(36) NOT NULL COMMENT '报销单唯一ID（UUID格式，全局唯一）',
  `user_id` VARCHAR(64) NOT NULL COMMENT '报销人ID（关联企业用户系统，此处简化为字符串）',
  `user_name` VARCHAR(64) NOT NULL COMMENT '报销人姓名',
  `dept_name` VARCHAR(64) NOT NULL COMMENT '报销人所属部门',
  `total_amount` DECIMAL(12,2) NOT NULL COMMENT '报销单总金额（单位：元，保留2位小数）',
  `category` VARCHAR(32) NOT NULL COMMENT '报销类目（枚举值：差旅费/办公费/业务招待费/其他）',
  `reason` TEXT NOT NULL COMMENT '报销事由（用户填写的详细说明，支持长文本）',
  `city_level` VARCHAR(16) DEFAULT '其他城市' COMMENT '出差城市级别（关联city_levels表，默认其他城市）',
  `status` VARCHAR(16) NOT NULL DEFAULT 'draft' COMMENT '报销单状态（枚举值：draft-草稿/uploaded-已上传/auditing-审核中/passed-审核通过/rejected-审核驳回/manual-需人工复核）',
  `upload_time` TIMESTAMP NOT NULL COMMENT '报销单上传时间（已上传状态时必填）',
  `ext_json` TEXT DEFAULT '{}' COMMENT '扩展信息（JSON格式字符串，如出差起止时间、项目编号等）',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`) COMMENT '按报销人ID查询索引，优化个人报销单查询',
  KEY `idx_status` (`status`) COMMENT '按状态查询索引，优化待审核/已审核单据筛选',
  KEY `idx_upload_time` (`upload_time`) COMMENT '按上传时间查询索引，优化时间范围查询',
  KEY `idx_user_status` (`user_id`, `status`) COMMENT '用户+状态联合索引，优化“查询某用户待审核单据”场景'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='报销单主表（核心业务主表）';

-- ----------------------------
-- 3. 发票表（invoices）
-- 作用：存储单张发票明细，与报销单为“一对多”关联（1张报销单可关联多张发票）
-- ----------------------------
DROP TABLE IF EXISTS `invoices`;
CREATE TABLE `invoices` (
  `id` VARCHAR(36) NOT NULL COMMENT '发票唯一ID（UUID格式）',
  `reimbursement_id` VARCHAR(36) NOT NULL COMMENT '关联报销单ID（外键关联reimbursements表）',
  `invoice_no` VARCHAR(64) NOT NULL COMMENT '发票号码（税务唯一标识，用于重复报销校验）',
  `invoice_type` VARCHAR(32) NOT NULL COMMENT '发票类型（枚举值：增值税专用发票/增值税普通发票/电子发票/其他）',
  `issuer_name` VARCHAR(128) NOT NULL COMMENT '开票方名称（OCR解析结果）',
  `issue_date` DATE NOT NULL COMMENT '开票日期（OCR解析结果，格式：YYYY-MM-DD）',
  `amount` DECIMAL(12,2) NOT NULL COMMENT '发票金额（不含税，单位：元）',
  `tax_amount` DECIMAL(12,2) DEFAULT 0.00 COMMENT '发票税额（单位：元，无税额填0）',
  `total_amount` DECIMAL(12,2) NOT NULL COMMENT '发票价税合计（=amount+tax_amount，单位：元）',
  `ocr_text` TEXT NOT NULL COMMENT '发票OCR完整解析文本（供大模型RAG检索使用）',
  `image_path` VARCHAR(255) NOT NULL COMMENT '发票图片存储路径（本地文件路径或MinIO访问URL）',
  `is_valid` TINYINT(1) DEFAULT 1 COMMENT '发票是否有效（1-有效，0-无效，OCR解析后初步判定）',
  `ocr_error_msg` VARCHAR(255) DEFAULT '' COMMENT 'OCR解析错误信息（如：图片模糊、识别不全）',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_invoice_no` (`invoice_no`) COMMENT '发票号码唯一约束，防止重复报销',
  KEY `idx_reimbursement_id` (`reimbursement_id`) COMMENT '关联报销单查询索引，优化发票-报销单关联查询',
  KEY `idx_issue_date` (`issue_date`) COMMENT '按开票日期查询索引，优化跨时间段发票筛选',
  -- 外键约束：删除报销单时同步删除关联发票（级联删除）
  CONSTRAINT `fk_invoice_reimbursement` FOREIGN KEY (`reimbursement_id`) REFERENCES `reimbursements` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='发票表（报销单关联表）';

-- ----------------------------
-- 4. 刚性规则表（audit_rules）
-- 作用：存储规则引擎（Grule）的刚性校验规则，支持动态加载与执行
-- ----------------------------
DROP TABLE IF EXISTS `audit_rules`;
CREATE TABLE `audit_rules` (
  `id` VARCHAR(36) NOT NULL COMMENT '规则唯一ID（UUID格式）',
  `rule_code` VARCHAR(32) NOT NULL COMMENT '规则编码（如：RULE_ACCOMMODATION_AMOUNT，全局唯一）',
  `rule_name` VARCHAR(128) NOT NULL COMMENT '规则名称（如：一线城市住宿费上限800元）',
  `rule_content` TEXT NOT NULL COMMENT '规则内容（Grule DSL语法，规则引擎执行的核心逻辑）',
  `priority` INT NOT NULL DEFAULT 5 COMMENT '规则优先级（1-10，1最高，决定规则执行顺序）',
  `category` VARCHAR(32) NOT NULL COMMENT '规则适用类目（枚举值：差旅费/办公费/所有类目/其他）',
  `status` VARCHAR(16) NOT NULL DEFAULT 'enabled' COMMENT '规则状态（枚举值：enabled-启用/disabled-禁用）',
  `description` TEXT DEFAULT '' COMMENT '规则描述（详细说明规则适用场景、校验逻辑）',
  `created_by` VARCHAR(64) NOT NULL COMMENT '规则创建人（财务管理员ID）',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_rule_code` (`rule_code`) COMMENT '规则编码唯一约束，避免规则冲突',
  KEY `idx_status` (`status`) COMMENT '按规则状态查询索引，优化启用规则筛选',
  KEY `idx_category` (`category`) COMMENT '按适用类目查询索引，优化规则匹配效率',
  KEY `idx_priority` (`priority`) COMMENT '按优先级查询索引，优化规则执行顺序排序',
  KEY `idx_category_status` (`category`, `status`) COMMENT '类目+状态联合索引，优化“按类目查询启用规则”场景'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='刚性规则表（Grule规则引擎核心表）';

-- ----------------------------
-- 5. 报销制度文档表（reimbursement_documents）
-- 作用：存储报销制度文档（原文+向量嵌入），支撑大模型RAG检索
-- 说明：MySQL 5.6无原生向量类型，向量嵌入转换为“JSON数组→BASE64字符串”存储
-- ----------------------------
DROP TABLE IF EXISTS `reimbursement_documents`;
CREATE TABLE `reimbursement_documents` (
  `id` VARCHAR(36) NOT NULL COMMENT '文档分片唯一ID（UUID格式）',
  `doc_name` VARCHAR(128) NOT NULL COMMENT '文档名称（如：2024年企业报销制度V1.0）',
  `doc_version` VARCHAR(16) NOT NULL COMMENT '文档版本（如：V1.0、V2.1，全局唯一）',
  `file_path` VARCHAR(255) NOT NULL COMMENT '原始文档存储路径（PDF/Word文件路径或MinIO访问URL）',
  `content` TEXT NOT NULL COMMENT '文档全文文本（解析后的纯文本，用于分片处理）',
  `chunk_id` VARCHAR(64) NOT NULL COMMENT '文本分片ID（如：DOC_V1.0_CHUNK_001，单文档按500字分片）',
  `chunk_content` TEXT NOT NULL COMMENT '分片文本内容（大模型RAG检索的核心上下文）',
  `embedding` TEXT NOT NULL COMMENT '向量嵌入（768维数组→JSON→BASE64编码字符串，业务层需解码）',
  `status` VARCHAR(16) NOT NULL DEFAULT 'valid' COMMENT '分片状态（枚举值：valid-有效/expired-过期）',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_doc_version_chunk` (`doc_version`, `chunk_id`) COMMENT '版本+分片ID唯一约束，避免重复分片',
  KEY `idx_doc_version` (`doc_version`) COMMENT '按文档版本查询索引，优化最新版本制度检索',
  KEY `idx_status` (`status`) COMMENT '按分片状态查询索引，优化有效分片筛选',
  KEY `idx_doc_version_status` (`doc_version`, `status`) COMMENT '版本+状态联合索引，优化“查询最新版本有效分片”场景'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='报销制度文档表（RAG检索核心表）';

-- ----------------------------
-- 6. 审核任务表（audit_tasks）
-- 作用：跟踪异步审核流程状态（大模型调用耗时较长时必备），避免重复触发审核
-- ----------------------------
DROP TABLE IF EXISTS `audit_tasks`;
CREATE TABLE `audit_tasks` (
  `id` VARCHAR(36) NOT NULL COMMENT '任务唯一ID（UUID格式）',
  `reimbursement_id` VARCHAR(36) NOT NULL COMMENT '关联报销单ID',
  `task_status` VARCHAR(16) NOT NULL DEFAULT 'pending' COMMENT '任务状态（枚举值：pending-待执行/processing-执行中/completed-完成/failed-失败）',
  `progress` INT NOT NULL DEFAULT 0 COMMENT '审核进度（0-100，如：规则校验50%→RAG检索80%→完成100%）',
  `error_msg` TEXT DEFAULT '' COMMENT '任务失败原因（如：大模型API调用超时、OCR解析失败）',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数（默认最多重试2次）',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '任务创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '任务更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_reimbursement_id` (`reimbursement_id`) COMMENT '报销单ID唯一约束，1张报销单对应1个审核任务',
  KEY `idx_task_status` (`task_status`) COMMENT '按任务状态查询索引，优化待执行/失败任务筛选',
  KEY `idx_retry_count` (`retry_count`) COMMENT '按重试次数查询索引，优化失败任务重试逻辑',
  KEY `idx_status_retry` (`task_status`, `retry_count`) COMMENT '状态+重试次数联合索引，优化“重试失败任务”场景',
  -- 外键约束：删除报销单时同步删除关联审核任务
  CONSTRAINT `fk_task_reimbursement` FOREIGN KEY (`reimbursement_id`) REFERENCES `reimbursements` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审核任务表（异步审核跟踪）';

-- ----------------------------
-- 7. 审核报告主表（audit_reports）
-- 作用：存储报销单最终审核结果（规则引擎+大模型RAG合并结果）
-- ----------------------------
DROP TABLE IF EXISTS `audit_reports`;
CREATE TABLE `audit_reports` (
  `id` VARCHAR(36) NOT NULL COMMENT '报告唯一ID（UUID格式）',
  `reimbursement_id` VARCHAR(36) NOT NULL COMMENT '关联报销单ID（1张报销单对应1份报告）',
  `rule_audit_result` TEXT NOT NULL COMMENT '规则引擎审核结果（JSON格式字符串，如：{"passed":false,"violations":[{"rule_code":"RULE_001","reason":"超金额上限","suggestion":"核减金额"}]}）',
  `rag_audit_result` TEXT NOT NULL COMMENT '大模型RAG审核结果（JSON格式字符串，如：{"compliant":true,"reason":"异地出差补贴符合制度第3条","suggestion":""}）',
  `final_conclusion` VARCHAR(16) NOT NULL COMMENT '最终结论（枚举值：passed-通过/rejected-驳回/manual-需人工复核）',
  `final_reason` TEXT NOT NULL COMMENT '最终结论说明（合并规则引擎+RAG的核心原因，支持长文本）',
  `audit_time` TIMESTAMP NOT NULL COMMENT '审核完成时间',
  `ext_json` TEXT DEFAULT '{}' COMMENT '扩展信息（JSON格式字符串，如大模型调用耗时、RAG检索命中分片ID）',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_reimbursement_report` (`reimbursement_id`) COMMENT '报销单ID唯一约束，避免重复生成报告',
  KEY `idx_final_conclusion` (`final_conclusion`) COMMENT '按最终结论查询索引，优化审核结果统计',
  KEY `idx_audit_time` (`audit_time`) COMMENT '按审核时间查询索引，优化时间范围报告筛选',
  -- 外键约束：删除报销单时同步删除关联审核报告
  CONSTRAINT `fk_report_reimbursement` FOREIGN KEY (`reimbursement_id`) REFERENCES `reimbursements` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审核报告主表（最终审核结果存储）';

-- ----------------------------
-- 8. OCR结果缓存表（ocr_caches）- 可选表
-- 作用：缓存发票OCR解析结果，提升重复上传发票的解析效率（简化版可省略）
-- ----------------------------
DROP TABLE IF EXISTS `ocr_caches`;
CREATE TABLE `ocr_caches` (
  `id` VARCHAR(36) NOT NULL COMMENT '缓存唯一ID（UUID格式）',
  `invoice_no` VARCHAR(64) NOT NULL COMMENT '发票号码（缓存KEY，关联发票表）',
  `ocr_text` TEXT NOT NULL COMMENT 'OCR解析结果（完整文本）',
  `expire_time` TIMESTAMP NOT NULL COMMENT '缓存过期时间（默认1小时，如：DATE_ADD(NOW(), INTERVAL 1 HOUR)）',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '缓存创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '缓存更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_invoice_no_cache` (`invoice_no`) COMMENT '发票号码唯一约束，避免重复缓存',
  KEY `idx_expire_time` (`expire_time`) COMMENT '按过期时间查询索引，用于定时清理过期缓存'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='OCR结果缓存表（可选，提升解析效率）';

-- 开启外键约束检查（恢复默认配置）
SET FOREIGN_KEY_CHECKS = 1;

-- ----------------------------
-- 初始化基础数据 - 城市级别配置（直接可用）
-- ----------------------------
INSERT INTO `city_levels` (`city_name`, `city_level`, `remark`) VALUES
('北京', '一线城市', '2024年第一财经榜单认定，首都'),
('上海', '一线城市', '2024年第一财经榜单认定，直辖市'),
('广州', '一线城市', '2024年第一财经榜单认定，省会'),
('深圳', '一线城市', '2024年第一财经榜单认定，经济特区'),
('杭州', '新一线城市', '2024年第一财经榜单认定'),
('成都', '新一线城市', '2024年第一财经榜单认定'),
('武汉', '新一线城市', '2024年第一财经榜单认定'),
('西安', '新一线城市', '2024年第一财经榜单认定'),
('南京', '新一线城市', '2024年第一财经榜单认定'),
('重庆', '新一线城市', '2024年第一财经榜单认定，直辖市'),
('天津', '二线城市', '2024年第一财经榜单认定，直辖市'),
('苏州', '二线城市', '2024年第一财经榜单认定'),
('郑州', '二线城市', '2024年第一财经榜单认定'),
('长沙', '二线城市', '2024年第一财经榜单认定'),
('青岛', '二线城市', '2024年第一财经榜单认定');

-- ----------------------------
-- 初始化基础数据 - 刚性规则示例（可根据实际业务调整）
-- 说明：MySQL 5.6 无原生 UUID()，手动生成示例 UUID（实际开发需通过业务层生成唯一UUID）
-- ----------------------------
INSERT INTO `audit_rules` (`id`, `rule_code`, `rule_name`, `rule_content`, `priority`, `category`, `status`, `description`, `created_by`) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'RULE_ACCOMMODATION_FIRST_TIER', '一线城市住宿费上限', 'Reimbursement.Category == "差旅费" && Reimbursement.CityLevel == "一线城市" && Reimbursement.TotalAmount > 800', 3, '差旅费', 'enabled', '一线城市出差住宿费单日上限800元，超出自动触发违规提示', 'admin'),
('550e8400-e29b-41d4-a716-446655440002', 'RULE_ACCOMMODATION_NEW_FIRST_TIER', '新一线城市住宿费上限', 'Reimbursement.Category == "差旅费" && Reimbursement.CityLevel == "新一线城市" && Reimbursement.TotalAmount > 600', 3, '差旅费', 'enabled', '新一线城市出差住宿费单日上限600元，超出自动触发违规提示', 'admin'),
('550e8400-e29b-41d4-a716-446655440003', 'RULE_ACCOMMODATION_SECOND_TIER', '二线城市住宿费上限', 'Reimbursement.Category == "差旅费" && Reimbursement.CityLevel == "二线城市" && Reimbursement.TotalAmount > 400', 3, '差旅费', 'enabled', '二线城市出差住宿费单日上限400元，超出自动触发违规提示', 'admin'),
('550e8400-e29b-41d4-a716-446655440004', 'RULE_MEAL_ALLOWANCE_DAILY', '单日餐饮补贴上限', 'Reimbursement.Category == "差旅费" && Reimbursement.TotalAmount > 150', 4, '差旅费', 'enabled', '出差期间单日餐饮补贴上限150元，超出自动触发违规提示', 'admin'),
('550e8400-e29b-41d4-a716-446655440005', 'RULE_RECEIPT_DUPLICATE', '发票重复报销校验', 'EXISTS(SELECT 1 FROM invoices WHERE invoice_no = CURRENT_INVOICE_NO AND reimbursement_id != CURRENT_REIMBURSEMENT_ID)', 2, '所有类目', 'enabled', '校验发票号码是否已在其他报销单中使用，防止重复报销', 'admin'),
('550e8400-e29b-41d4-a716-446655440006', 'RULE_BUSINESS_ENTERTAINMENT_LIMIT', '业务招待费上限', 'Reimbursement.Category == "业务招待费" && Reimbursement.TotalAmount > 2000', 2, '业务招待费', 'enabled', '单次业务招待费上限2000元，超需人工复核', 'admin'),
('550e8400-e29b-41d4-a716-446655440007', 'RULE_OFFICE_SUPPLIES_FREQUENCY', '办公费月报销频次', 'Reimbursement.Category == "办公费" && (SELECT COUNT(1) FROM reimbursements WHERE user_id = CURRENT_USER_ID AND category = "办公费" AND DATE_FORMAT(upload_time, "%Y-%m") = DATE_FORMAT(NOW(), "%Y-%m")) > 5', 4, '办公费', 'enabled', '同一用户每月办公费报销次数不超过5次，超次自动触发违规', 'admin');

-- ------------------------------------------------------
-- 执行说明（MySQL 5.6 专属）
-- ------------------------------------------------------
-- 1. 执行前置条件：
--    - 确保数据库编码为 utf8mb4（支持emoji和特殊字符），执行：ALTER DATABASE 数据库名 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
--    - 开启 InnoDB 引擎（默认开启，无需额外配置）
-- 2. 执行方式：
--    登录MySQL终端：mysql -u用户名 -p密码
--    切换数据库：USE 目标数据库名;
--    执行脚本：SOURCE /路径/reimbursement-audit.sql;
-- 3. 关键适配点说明：
--    - UUID生成：业务层需通过工具生成唯一UUID（如Go的 github.com/google/uuid 包），避免手动输入重复
--    - JSON处理：业务层需将 TEXT 类型的 JSON 字符串序列化为结构体（如Go的 encoding/json 包）
--    - 向量解码：embedding字段的BASE64字符串需先解码为JSON字符串，再转换为向量数组（如Go的 encoding/base64 + encoding/json 包）
--    - 时间函数：MySQL 5.6 中 DATE_ADD 函数可用，缓存过期时间建议用 DATE_ADD(NOW(), INTERVAL 1 HOUR)
-- 4. 性能优化：
--    - 数据量超10万条时，对 reimbursements、invoices表按 upload_time 字段分区（如按季度分区）
--    - 定期清理OCR过期缓存：执行 DELETE FROM ocr_caches WHERE expire_time < NOW();（可通过定时任务实现）
-- 5. 安全建议：
--    - 生产环境需对 user_id、dept_name 等敏感字段加密存储（如AES加密）
--    - 限制数据库用户权限，仅给应用程序分配 SELECT/INSERT/UPDATE/DELETE 权限，禁止 DROP/ALTER 等高危操作