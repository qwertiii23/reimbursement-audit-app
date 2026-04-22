-- 加班费标准表
CREATE TABLE IF NOT EXISTS overtime_standard (
    id VARCHAR(36) PRIMARY KEY COMMENT '主键ID',
    employee_level VARCHAR(50) NOT NULL COMMENT '员工职级：普通员工、部门经理、高管',
    hourly_rate DECIMAL(10, 2) NOT NULL COMMENT '时薪（元/小时）',
    daily_max_hours DECIMAL(10, 2) NOT NULL COMMENT '每日最长加班小时数',
    monthly_max_hours DECIMAL(10, 2) NOT NULL COMMENT '每月最长加班小时数',
    description TEXT COMMENT '标准说明',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_employee_level (employee_level),
    UNIQUE KEY uk_employee_level (employee_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='加班费标准表';

-- 插入加班费标准数据
INSERT INTO overtime_standard (id, employee_level, hourly_rate, daily_max_hours, monthly_max_hours, description) VALUES
('os-1', '普通员工', 50.00, 4.00, 36.00, '普通员工加班费标准：时薪50元/小时，每日最长4小时，每月最长36小时'),
('os-2', '部门经理', 80.00, 4.00, 36.00, '部门经理加班费标准：时薪80元/小时，每日最长4小时，每月最长36小时'),
('os-3', '高管', 120.00, 4.00, 36.00, '高管加班费标准：时薪120元/小时，每日最长4小时，每月最长36小时');
