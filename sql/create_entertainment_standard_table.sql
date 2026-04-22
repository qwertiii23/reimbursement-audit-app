-- 招待费标准表
CREATE TABLE IF NOT EXISTS entertainment_standard (
    id VARCHAR(36) PRIMARY KEY COMMENT '主键ID',
    guest_type VARCHAR(50) NOT NULL COMMENT '招待对象类型：client(客户)、partner(合作伙伴)、government(政府)、internal(内部)、other(其他)',
    employee_level VARCHAR(50) NOT NULL COMMENT '员工职级：普通员工、部门经理、高管',
    per_person_limit DECIMAL(10, 2) NOT NULL COMMENT '人均费用上限（元）',
    daily_limit DECIMAL(10, 2) NOT NULL COMMENT '每日费用上限（元）',
    description TEXT COMMENT '标准说明',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_guest_type (guest_type),
    INDEX idx_employee_level (employee_level),
    UNIQUE KEY uk_guest_level (guest_type, employee_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='招待费标准表';

-- 插入招待费标准数据
INSERT INTO entertainment_standard (id, guest_type, employee_level, per_person_limit, daily_limit, description) VALUES
-- 客户
('es-client-1', 'client', '普通员工', 200.00, 500.00, '普通员工招待客户标准'),
('es-client-2', 'client', '部门经理', 300.00, 800.00, '部门经理招待客户标准'),
('es-client-3', 'client', '高管', 500.00, 1500.00, '高管招待客户标准'),

-- 合作伙伴
('es-partner-1', 'partner', '普通员工', 150.00, 400.00, '普通员工招待合作伙伴标准'),
('es-partner-2', 'partner', '部门经理', 250.00, 600.00, '部门经理招待合作伙伴标准'),
('es-partner-3', 'partner', '高管', 400.00, 1200.00, '高管招待合作伙伴标准'),

-- 政府
('es-government-1', 'government', '普通员工', 250.00, 600.00, '普通员工招待政府标准'),
('es-government-2', 'government', '部门经理', 400.00, 1000.00, '部门经理招待政府标准'),
('es-government-3', 'government', '高管', 600.00, 2000.00, '高管招待政府标准'),

-- 内部
('es-internal-1', 'internal', '普通员工', 100.00, 300.00, '普通员工招待内部标准'),
('es-internal-2', 'internal', '部门经理', 150.00, 400.00, '部门经理招待内部标准'),
('es-internal-3', 'internal', '高管', 200.00, 500.00, '高管招待内部标准'),

-- 其他
('es-other-1', 'other', '普通员工', 100.00, 300.00, '普通员工招待其他标准'),
('es-other-2', 'other', '部门经理', 150.00, 400.00, '部门经理招待其他标准'),
('es-other-3', 'other', '高管', 200.00, 500.00, '高管招待其他标准');
