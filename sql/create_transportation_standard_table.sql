-- 交通费标准表
CREATE TABLE IF NOT EXISTS transportation_standard (
    id VARCHAR(36) PRIMARY KEY COMMENT '主键ID',
    transport_type VARCHAR(50) NOT NULL COMMENT '出行方式：plane(飞机)、high_speed_rail(高铁)、train(火车)、car(汽车)、bus(大巴)',
    employee_level VARCHAR(50) NOT NULL COMMENT '员工职级：普通员工、部门经理、高管',
    daily_limit DECIMAL(10, 2) NOT NULL COMMENT '每日费用上限（元）',
    single_trip_limit DECIMAL(10, 2) NOT NULL COMMENT '单次行程费用上限（元）',
    description TEXT COMMENT '标准说明',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_transport_type (transport_type),
    INDEX idx_employee_level (employee_level),
    UNIQUE KEY uk_transport_level (transport_type, employee_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交通费标准表';

-- 插入交通费标准数据
INSERT INTO transportation_standard (id, transport_type, employee_level, daily_limit, single_trip_limit, description) VALUES
-- 飞机
('ts-plane-1', 'plane', '普通员工', 2000.00, 2000.00, '普通员工飞机票：经济舱，单次不超过2000元'),
('ts-plane-2', 'plane', '部门经理', 4000.00, 4000.00, '部门经理飞机票：商务舱，单次不超过4000元'),
('ts-plane-3', 'plane', '高管', 8000.00, 8000.00, '高管飞机票：头等舱，单次不超过8000元'),

-- 高铁
('ts-hsr-1', 'high_speed_rail', '普通员工', 800.00, 800.00, '普通员工高铁票：二等座，单次不超过800元'),
('ts-hsr-2', 'high_speed_rail', '部门经理', 1500.00, 1500.00, '部门经理高铁票：一等座，单次不超过1500元'),
('ts-hsr-3', 'high_speed_rail', '高管', 2000.00, 2000.00, '高管高铁票：商务座，单次不超过2000元'),

-- 火车
('ts-train-1', 'train', '普通员工', 500.00, 500.00, '普通员工火车票：硬座/硬卧，单次不超过500元'),
('ts-train-2', 'train', '部门经理', 800.00, 800.00, '部门经理火车票：软卧，单次不超过800元'),
('ts-train-3', 'train', '高管', 1000.00, 1000.00, '高管火车票：软卧，单次不超过1000元'),

-- 汽车
('ts-car-1', 'car', '普通员工', 1000.00, 500.00, '普通员工汽车：每日不超过1000元，单次不超过500元'),
('ts-car-2', 'car', '部门经理', 1500.00, 800.00, '部门经理汽车：每日不超过1500元，单次不超过800元'),
('ts-car-3', 'car', '高管', 2000.00, 1000.00, '高管汽车：每日不超过2000元，单次不超过1000元'),

-- 大巴
('ts-bus-1', 'bus', '普通员工', 300.00, 300.00, '普通员工大巴：单次不超过300元'),
('ts-bus-2', 'bus', '部门经理', 500.00, 500.00, '部门经理大巴：单次不超过500元'),
('ts-bus-3', 'bus', '高管', 800.00, 800.00, '高管大巴：单次不超过800元');
