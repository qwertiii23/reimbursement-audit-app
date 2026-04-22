-- =====================================================
-- 智能报销审核系统 - PostgreSQL 完整建表脚本
-- 适用于 PostgreSQL 13+
-- 注意: 需要先启用 pgvector 扩展
-- =====================================================

-- 启用扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgvector";

-- =====================================================
-- 1. 用户表 (users)
-- =====================================================
DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE users (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  username VARCHAR(50) NOT NULL,
  password VARCHAR(255) NOT NULL,
  email VARCHAR(100),
  real_name VARCHAR(50),
  role VARCHAR(20) NOT NULL DEFAULT 'user',
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  last_login TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uk_users_username ON users(username);
CREATE UNIQUE INDEX uk_users_email ON users(email);

-- =====================================================
-- 2. 报销单主表 (reimbursements)
-- =====================================================
DROP TABLE IF EXISTS reimbursements CASCADE;
CREATE TABLE reimbursements (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  user_id VARCHAR(36) NOT NULL,
  user_name VARCHAR(100) NOT NULL,
  department VARCHAR(100),
  applicant_level VARCHAR(20),
  type VARCHAR(50),
  title VARCHAR(200) NOT NULL,
  description TEXT,
  total_amount DECIMAL(10,2) NOT NULL,
  currency VARCHAR(10) DEFAULT 'CNY',
  apply_date DATE NOT NULL,
  expense_date DATE,
  start_date DATE,
  end_date DATE,
  destination VARCHAR(100),
  city VARCHAR(50),
  province VARCHAR(50),
  travel_reason VARCHAR(200),
  transportation VARCHAR(50),
  project_code VARCHAR(50),
  budget_code VARCHAR(50),
  approval_required BOOLEAN DEFAULT FALSE,
  approved_by VARCHAR(36),
  approved_at TIMESTAMP,
  audit_id VARCHAR(36),
  status VARCHAR(20) NOT NULL DEFAULT 'pending_submission',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_reimbursements_user_id ON reimbursements(user_id);
CREATE INDEX idx_reimbursements_status ON reimbursements(status);
CREATE INDEX idx_reimbursements_audit_id ON reimbursements(audit_id);

-- =====================================================
-- 3. 发票表 (invoices)
-- =====================================================
DROP TABLE IF EXISTS invoices CASCADE;
CREATE TABLE invoices (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  reimbursement_id VARCHAR(36) NOT NULL,
  type VARCHAR(50),
  code VARCHAR(50),
  number VARCHAR(50),
  date DATE,
  amount DECIMAL(10,2) NOT NULL,
  tax_amount DECIMAL(10,2),
  payer VARCHAR(100),
  payee VARCHAR(100),
  buyer_name VARCHAR(100),
  buyer_tax_no VARCHAR(50),
  seller_name VARCHAR(100),
  seller_tax_no VARCHAR(50),
  commodity_name VARCHAR(200),
  specification VARCHAR(100),
  unit VARCHAR(20),
  quantity DECIMAL(10,2),
  price DECIMAL(10,2),
  image_path VARCHAR(500),
  ocr_result TEXT,
  items JSONB,
  total_items INTEGER DEFAULT 0,
  main_commodity VARCHAR(200),
  status VARCHAR(20) NOT NULL DEFAULT 'pending_recognition',
  category VARCHAR(50),
  sub_category VARCHAR(50),
  expense_type VARCHAR(50),
  payment_method VARCHAR(50),
  merchant_type VARCHAR(50),
  merchant_code VARCHAR(50),
  location VARCHAR(100),
  city VARCHAR(50),
  province VARCHAR(50),
  country VARCHAR(50) DEFAULT 'China',
  purpose VARCHAR(200),
  project_code VARCHAR(50),
  department_code VARCHAR(50),
  cost_center VARCHAR(50),
  contract_number VARCHAR(50),
  approval_level VARCHAR(20),
  is_reimbursable BOOLEAN DEFAULT TRUE,
  is_personal BOOLEAN DEFAULT FALSE,
  is_vat BOOLEAN DEFAULT FALSE,
  vat_rate DECIMAL(5,2) DEFAULT 0.00,
  exchange_rate DECIMAL(10,4) DEFAULT 1.0000,
  original_amount DECIMAL(10,2) DEFAULT 0.00,
  original_currency VARCHAR(10),
  receipt_number VARCHAR(50),
  invoice_series VARCHAR(50),
  batch_number VARCHAR(50),
  valid_from DATE,
  valid_to DATE,
  is_electronic BOOLEAN DEFAULT FALSE,
  is_duplicate BOOLEAN DEFAULT FALSE,
  related_invoice_id VARCHAR(36),
  verification_status VARCHAR(20) DEFAULT 'unverified',
  verification_time TIMESTAMP,
  remarks TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_invoices_reimbursement_id ON invoices(reimbursement_id);
CREATE INDEX idx_invoices_number ON invoices(number);
CREATE INDEX idx_invoices_status ON invoices(status);
ALTER TABLE invoices ADD CONSTRAINT fk_invoices_reimbursement
  FOREIGN KEY (reimbursement_id) REFERENCES reimbursements(id) ON DELETE CASCADE;

-- =====================================================
-- 4. 审核结果表 (audit_results)
-- =====================================================
DROP TABLE IF EXISTS audit_results CASCADE;
CREATE TABLE audit_results (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  reimbursement_id VARCHAR(36) NOT NULL,
  status VARCHAR(20) NOT NULL,
  workflow_status VARCHAR(50),
  rule_pass BOOLEAN DEFAULT FALSE,
  rag_pass BOOLEAN DEFAULT FALSE,
  final_pass BOOLEAN DEFAULT FALSE,
  risk_level VARCHAR(20),
  risk_score DECIMAL(5,4) DEFAULT 0,
  reason TEXT,
  rule_results JSONB,
  rag_results JSONB,
  suggestions TEXT,
  started_at TIMESTAMP,
  completed_at TIMESTAMP,
  duration BIGINT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_audit_results_reimbursement_id ON audit_results(reimbursement_id);
CREATE INDEX idx_audit_results_status ON audit_results(status);

-- =====================================================
-- 5. 审核流程日志表 (audit_flow_logs)
-- =====================================================
DROP TABLE IF EXISTS audit_flow_logs CASCADE;
CREATE TABLE audit_flow_logs (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  reimbursement_id VARCHAR(36) NOT NULL,
  audit_id VARCHAR(36) NOT NULL,
  flow_status VARCHAR(50),
  flow_type VARCHAR(20),
  operator_id VARCHAR(36),
  operator_name VARCHAR(100),
  action VARCHAR(50),
  reason TEXT,
  result TEXT,
  ip_address VARCHAR(50),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_audit_flow_logs_reimbursement_id ON audit_flow_logs(reimbursement_id);
CREATE INDEX idx_audit_flow_logs_audit_id ON audit_flow_logs(audit_id);

-- =====================================================
-- 6. 规则引擎-规则表 (rule_engine_rules)
-- =====================================================
DROP TABLE IF EXISTS rule_engine_rules CASCADE;
CREATE TABLE rule_engine_rules (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  name VARCHAR(100) NOT NULL,
  description TEXT,
  priority INTEGER DEFAULT 0,
  enabled BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_rule_engine_rules_enabled ON rule_engine_rules(enabled);

-- =====================================================
-- 7. 规则引擎-特征表 (features)
-- =====================================================
DROP TABLE IF EXISTS features CASCADE;
CREATE TABLE features (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  name VARCHAR(100) NOT NULL,
  code VARCHAR(50) NOT NULL,
  description TEXT,
  type VARCHAR(20) NOT NULL,
  value_type VARCHAR(20) NOT NULL,
  category VARCHAR(50),
  enabled BOOLEAN DEFAULT TRUE,
  function_name VARCHAR(100),
  function_config JSONB,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uk_features_name ON features(name);
CREATE UNIQUE INDEX uk_features_code ON features(code);
CREATE INDEX idx_features_category ON features(category);
CREATE INDEX idx_features_enabled ON features(enabled);

-- =====================================================
-- 8. 规则引擎-条件表 (conditions)
-- =====================================================
DROP TABLE IF EXISTS conditions CASCADE;
CREATE TABLE conditions (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  rule_id VARCHAR(36) NOT NULL,
  feature_id VARCHAR(36) NOT NULL,
  operator VARCHAR(20) NOT NULL,
  value TEXT,
  logic_op VARCHAR(10) DEFAULT 'and',
  sort_order INTEGER DEFAULT 0
);
CREATE INDEX idx_conditions_rule_id ON conditions(rule_id);
CREATE INDEX idx_conditions_feature_id ON conditions(feature_id);
ALTER TABLE conditions ADD CONSTRAINT fk_conditions_rule
  FOREIGN KEY (rule_id) REFERENCES rule_engine_rules(id) ON DELETE CASCADE;
ALTER TABLE conditions ADD CONSTRAINT fk_conditions_feature
  FOREIGN KEY (feature_id) REFERENCES features(id) ON DELETE CASCADE;

-- =====================================================
-- 9. 规则引擎-决策表 (decisions)
-- =====================================================
DROP TABLE IF EXISTS decisions CASCADE;
CREATE TABLE decisions (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  rule_id VARCHAR(36) NOT NULL UNIQUE,
  type VARCHAR(20) NOT NULL,
  reason TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE decisions ADD CONSTRAINT fk_decisions_rule
  FOREIGN KEY (rule_id) REFERENCES rule_engine_rules(id) ON DELETE CASCADE;

-- =====================================================
-- 10. 规则引擎-特征值表 (feature_values)
-- =====================================================
DROP TABLE IF EXISTS feature_values CASCADE;
CREATE TABLE feature_values (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  feature_id VARCHAR(36) NOT NULL,
  value VARCHAR(255) NOT NULL,
  label VARCHAR(255) NOT NULL,
  sort_order INTEGER DEFAULT 0,
  enabled BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_feature_values_feature_id ON feature_values(feature_id);
ALTER TABLE feature_values ADD CONSTRAINT fk_feature_values_feature
  FOREIGN KEY (feature_id) REFERENCES features(id) ON DELETE CASCADE;

-- =====================================================
-- 11. 知识库文件表 (knowledge_files)
-- =====================================================
DROP TABLE IF EXISTS knowledge_files CASCADE;
CREATE TABLE knowledge_files (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  file_name VARCHAR(255) NOT NULL,
  file_path VARCHAR(500) NOT NULL,
  file_type VARCHAR(50) NOT NULL,
  file_size BIGINT,
  category VARCHAR(100),
  description TEXT,
  uploaded_by VARCHAR(36),
  uploader_name VARCHAR(100),
  download_count INTEGER DEFAULT 0,
  status VARCHAR(20) DEFAULT 'active',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_knowledge_files_category ON knowledge_files(category);
CREATE INDEX idx_knowledge_files_status ON knowledge_files(status);

-- =====================================================
-- 12. 城市级别配置表 (city_tiers)
-- =====================================================
DROP TABLE IF EXISTS city_tiers CASCADE;
CREATE TABLE city_tiers (
  id SERIAL PRIMARY KEY,
  city_name VARCHAR(64) NOT NULL,
  city_level VARCHAR(16) NOT NULL,
  remark VARCHAR(255) DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uk_city_tiers_city_name ON city_tiers(city_name);

-- =====================================================
-- 13. 住宿费标准表 (accommodation_standards)
-- =====================================================
DROP TABLE IF EXISTS accommodation_standards CASCADE;
CREATE TABLE accommodation_standards (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  city_level VARCHAR(16) NOT NULL,
  star_rating VARCHAR(20) NOT NULL,
  standard_amount DECIMAL(10,2) NOT NULL,
  effective_date DATE NOT NULL,
  expiry_date DATE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_accommodation_standards_city_level ON accommodation_standards(city_level);

-- =====================================================
-- 14. 餐饮费标准表 (meal_standards)
-- =====================================================
DROP TABLE IF EXISTS meal_standards CASCADE;
CREATE TABLE meal_standards (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  city_level VARCHAR(16) NOT NULL,
  meal_type VARCHAR(20) NOT NULL,
  standard_amount DECIMAL(10,2) NOT NULL,
  effective_date DATE NOT NULL,
  expiry_date DATE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_meal_standards_city_level ON meal_standards(city_level);

-- =====================================================
-- 15. 招待费标准表 (entertainment_standards)
-- =====================================================
DROP TABLE IF EXISTS entertainment_standards CASCADE;
CREATE TABLE entertainment_standards (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  city_level VARCHAR(16) NOT NULL,
  entertainment_type VARCHAR(20) NOT NULL,
  standard_amount DECIMAL(10,2) NOT NULL,
  effective_date DATE NOT NULL,
  expiry_date DATE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_entertainment_standards_city_level ON entertainment_standards(city_level);

-- =====================================================
-- 16. 加班费标准表 (overtime_standards)
-- =====================================================
DROP TABLE IF EXISTS overtime_standards CASCADE;
CREATE TABLE overtime_standards (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  meal_standard DECIMAL(10,2) NOT NULL,
  transport_allowance DECIMAL(10,2) NOT NULL,
  effective_date DATE NOT NULL,
  expiry_date DATE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- 17. 交通费标准表 (transportation_standards)
-- =====================================================
DROP TABLE IF EXISTS transportation_standards CASCADE;
CREATE TABLE transportation_standards (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  transport_type VARCHAR(50) NOT NULL,
  city_level VARCHAR(16),
  distance_range VARCHAR(50),
  standard_amount DECIMAL(10,2) NOT NULL,
  effective_date DATE NOT NULL,
  expiry_date DATE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_transportation_standards_transport_type ON transportation_standards(transport_type);

-- =====================================================
-- 18. 数据库迁移记录表 (migrations)
-- =====================================================
DROP TABLE IF EXISTS migrations CASCADE;
CREATE TABLE migrations (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  name VARCHAR(255) NOT NULL,
  applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- 19. RAG向量存储表 (document_chunks) - 用于RAG智能审核
-- =====================================================
DROP TABLE IF EXISTS document_chunks CASCADE;
CREATE TABLE document_chunks (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  document_id VARCHAR(36) NOT NULL,
  chunk_index INTEGER NOT NULL,
  content TEXT NOT NULL,
  embedding VECTOR(1536),
  metadata JSONB,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_document_chunks_document_id ON document_chunks(document_id);
CREATE INDEX idx_document_chunks_embedding ON document_chunks USING ivfflat (embedding vector_cosine_ops);

-- =====================================================
-- 20. RAG文档表 (rag_documents)
-- =====================================================
DROP TABLE IF EXISTS rag_documents CASCADE;
CREATE TABLE rag_documents (
  id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::VARCHAR(36),
  title VARCHAR(255) NOT NULL,
  content TEXT,
  source_type VARCHAR(50),
  source_id VARCHAR(36),
  category VARCHAR(100),
  tags TEXT[],
  metadata JSONB,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_rag_documents_category ON rag_documents(category);
CREATE INDEX idx_rag_documents_source_type ON rag_documents(source_type);

-- =====================================================
-- 初始化基础数据
-- =====================================================

-- 插入管理员用户 (密码: admin123)
INSERT INTO users (id, username, password, email, real_name, role, status) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin@example.com', '系统管理员', 'admin', 'active');

-- 插入财务用户 (密码: finance123)
INSERT INTO users (id, username, password, email, real_name, role, status) VALUES
('550e8400-e29b-41d4-a716-446655440002', 'finance', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'finance@example.com', '财务人员', 'finance', 'active');

-- 插入测试用户 (密码: user123)
INSERT INTO users (id, username, password, email, real_name, role, status) VALUES
('550e8400-e29b-41d4-a716-446655440003', 'user', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'user@example.com', '测试用户', 'user', 'active');

-- 插入城市级别配置
INSERT INTO city_tiers (city_name, city_level, remark) VALUES
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

-- 插入特征数据
INSERT INTO features (id, name, code, description, type, value_type, category, enabled, function_name, function_config) VALUES
('detect-photoshop-feature', '是否P图', 'is_photoshopped', '检测发票图片是否经过P图处理', 'boolean', 'single', 'image', true, 'detect_photoshop', '{}'),
('3433273d-626f-4374-8819-b661620b2f4a', '报销单金额', 'reimbursement_amount', '报销单总金额', 'number', 'single', 'amount', true, 'reimbursement_total_amount', '{}'),
('3cb0c294-3da9-11f1-bd57-cf7f11d80696', '报销总金额', 'reimbursement_total_amount', '报销单总金额', 'number', 'single', 'amount', true, 'reimbursement_total_amount', '{}'),
('626addbc-3da9-11f1-bd57-cf7f11d80696', '发票距今天数', 'invoice_days_from_today', '发票日期距今天的天数', 'number', 'single', 'time', true, 'invoice_days_from_today', '{}'),
('626b0936-3da9-11f1-bd57-cf7f11d80696', '出差天数', 'trip_duration', '出差起止日期之间的天数', 'number', 'single', 'time', true, 'trip_duration', '{}'),
('626b1de0-3da9-11f1-bd57-cf7f11d80696', '发票类型', 'invoice_type', '发票的类型', 'string', 'single', 'invoice', true, 'invoice_type', '{}'),
('626b31fe-3da9-11f1-bd57-cf7f11d80696', '商品名称', 'commodity_name', '发票商品名称', 'string', 'single', 'invoice', true, 'commodity_name', '{}'),
('626b3cb2-3da9-11f1-bd57-cf7f11d80696', '报销类型', 'reimbursement_type', '报销单类型', 'string', 'single', 'reimbursement', true, 'reimbursement_type', '{}'),
('626b446e-3da9-11f1-bd57-cf7f11d80696', '发票单价', 'invoice_price', '发票商品单价', 'number', 'single', 'amount', true, 'invoice_price', '{}'),
('4f99ee85-8817-4989-b162-4496eb15bc64', '发票分类', 'invoice_category', '发票分类', 'string', 'single', 'invoice', true, NULL, '{}'),
('7184644e-7be8-46d8-8a2c-54769982d0b0', '发票子分类', 'invoice_subcategory', '发票子分类', 'string', 'single', 'invoice', true, NULL, '{}'),
('a63c79a1-902d-4c94-b8fa-d69784bb4761', '发票金额', 'invoice_amount', '发票金额', 'number', 'single', 'amount', true, 'invoice_amount', '{}'),
('feat-invoice-fraud-detection', '发票舞弊检测', 'invoice_fraud_detection', '检测发票是否存在舞弊风险', 'boolean', 'single', 'fraud', true, 'invoice_fraud_detection', '{}'),
('feat-invoice-code-length', '发票代码长度', 'invoice_code_length', '检测发票代码长度是否合规', 'boolean', 'single', 'format', true, 'invoice_code_length', '{}'),
('feat-invoice-type-validation', '发票类型校验', 'invoice_type_validation', '校验发票类型是否符合报销要求', 'boolean', 'single', 'validation', true, 'invoice_type_validation', '{}'),
('feat-invoice-amount-range', '发票金额范围', 'invoice_amount_range', '校验发票金额是否在合理范围内', 'boolean', 'single', 'validation', true, 'invoice_amount_range', '{}'),
('feat-invoice-date-validity', '开票日期有效性', 'invoice_date_validity', '校验开票日期是否有效', 'boolean', 'single', 'validation', true, 'invoice_date_validity', '{}'),
('feat-invoice-number-format', '发票号码格式', 'invoice_number_format', '校验发票号码格式', 'boolean', 'single', 'format', true, 'invoice_number_format', '{}'),
('feat-product-name-compliance', '商品名称合规', 'product_name_compliance', '校验商品名称是否合规', 'boolean', 'single', 'validation', true, 'product_name_compliance', '{}'),
('feat-invoice-duplicate-check', '发票重复性校验', 'invoice_duplicate_check', '检测发票是否重复报销', 'boolean', 'single', 'duplicate', true, 'invoice_duplicate_check', '{}'),
('feat-smart-accommodation', '智能住宿费校验', 'smart_accommodation_validation', '根据城市级别智能校验住宿费', 'boolean', 'single', 'validation', true, 'smart_accommodation_validation', '{}'),
('feat-transportation', '交通费校验', 'transportation_validation', '校验交通费是否超标', 'boolean', 'single', 'validation', true, 'transportation_validation', '{}'),
('feat-meal', '餐饮费校验', 'meal_validation', '校验餐饮费是否超标', 'boolean', 'single', 'validation', true, 'meal_validation', '{}'),
('feat-entertainment', '招待费校验', 'entertainment_validation', '校验招待费是否超标', 'boolean', 'single', 'validation', true, 'entertainment_validation', '{}'),
('feat-trip-duration', '差旅天数校验', 'trip_duration_validation', '校验出差天数是否合理', 'boolean', 'single', 'validation', true, 'trip_duration_validation', '{}'),
('feat-overtime', '加班费校验', 'overtime_validation', '校验加班费是否合规', 'boolean', 'single', 'validation', true, 'overtime_validation', '{}'),
('feat-merchant-type-validation', '商户类型校验', 'merchant_type_validation', '校验商户类型是否合规', 'boolean', 'single', 'validation', true, 'merchant_type_validation', '{}'),
('feat-invoice-code-number-validation', '发票代码号码校验', 'invoice_code_number_validation', '校验发票代码号码是否正确', 'boolean', 'single', 'format', true, 'invoice_code_number_validation', '{}');

-- =====================================================
-- 初始化加班费标准
-- =====================================================
INSERT INTO overtime_standards (id, meal_standard, transport_allowance, effective_date) VALUES
(uuid_generate_v4()::VARCHAR(36), 50.00, 50.00, '2024-01-01');
