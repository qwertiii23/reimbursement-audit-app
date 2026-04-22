-- 餐饮费标准表
CREATE TABLE IF NOT EXISTS meal_standard (
    id VARCHAR(36) PRIMARY KEY COMMENT '主键ID',
    city_tier INT NOT NULL COMMENT '城市等级：1(一线城市)、2(二线城市)、3(三线城市)',
    employee_level VARCHAR(50) NOT NULL COMMENT '员工职级：普通员工、部门经理、高管',
    meal_type VARCHAR(50) NOT NULL COMMENT '用餐类型：breakfast(早餐)、lunch(午餐)、dinner(晚餐)',
    daily_limit DECIMAL(10, 2) NOT NULL COMMENT '每日费用上限（元）',
    per_meal_limit DECIMAL(10, 2) NOT NULL COMMENT '单次用餐费用上限（元）',
    description TEXT COMMENT '标准说明',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_city_tier (city_tier),
    INDEX idx_employee_level (employee_level),
    INDEX idx_meal_type (meal_type),
    UNIQUE KEY uk_city_level_meal (city_tier, employee_level, meal_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='餐饮费标准表';

-- 插入餐饮费标准数据
INSERT INTO meal_standard (id, city_tier, employee_level, meal_type, daily_limit, per_meal_limit, description) VALUES
-- 一线城市
('ms-1-1-breakfast', 1, '普通员工', 'breakfast', 100.00, 50.00, '一线城市普通员工早餐标准'),
('ms-1-1-lunch', 1, '普通员工', 'lunch', 150.00, 80.00, '一线城市普通员工午餐标准'),
('ms-1-1-dinner', 1, '普通员工', 'dinner', 200.00, 100.00, '一线城市普通员工晚餐标准'),

('ms-1-2-breakfast', 1, '部门经理', 'breakfast', 150.00, 80.00, '一线城市部门经理早餐标准'),
('ms-1-2-lunch', 1, '部门经理', 'lunch', 200.00, 100.00, '一线城市部门经理午餐标准'),
('ms-1-2-dinner', 1, '部门经理', 'dinner', 300.00, 150.00, '一线城市部门经理晚餐标准'),

('ms-1-3-breakfast', 1, '高管', 'breakfast', 200.00, 100.00, '一线城市高管早餐标准'),
('ms-1-3-lunch', 1, '高管', 'lunch', 300.00, 150.00, '一线城市高管午餐标准'),
('ms-1-3-dinner', 1, '高管', 'dinner', 400.00, 200.00, '一线城市高管晚餐标准'),

-- 二线城市
('ms-2-1-breakfast', 2, '普通员工', 'breakfast', 80.00, 40.00, '二线城市普通员工早餐标准'),
('ms-2-1-lunch', 2, '普通员工', 'lunch', 120.00, 60.00, '二线城市普通员工午餐标准'),
('ms-2-1-dinner', 2, '普通员工', 'dinner', 160.00, 80.00, '二线城市普通员工晚餐标准'),

('ms-2-2-breakfast', 2, '部门经理', 'breakfast', 120.00, 60.00, '二线城市部门经理早餐标准'),
('ms-2-2-lunch', 2, '部门经理', 'lunch', 160.00, 80.00, '二线城市部门经理午餐标准'),
('ms-2-2-dinner', 2, '部门经理', 'dinner', 240.00, 120.00, '二线城市部门经理晚餐标准'),

('ms-2-3-breakfast', 2, '高管', 'breakfast', 160.00, 80.00, '二线城市高管早餐标准'),
('ms-2-3-lunch', 2, '高管', 'lunch', 240.00, 120.00, '二线城市高管午餐标准'),
('ms-2-3-dinner', 2, '高管', 'dinner', 320.00, 160.00, '二线城市高管晚餐标准'),

-- 三线城市
('ms-3-1-breakfast', 3, '普通员工', 'breakfast', 60.00, 30.00, '三线城市普通员工早餐标准'),
('ms-3-1-lunch', 3, '普通员工', 'lunch', 90.00, 45.00, '三线城市普通员工午餐标准'),
('ms-3-1-dinner', 3, '普通员工', 'dinner', 120.00, 60.00, '三线城市普通员工晚餐标准'),

('ms-3-2-breakfast', 3, '部门经理', 'breakfast', 90.00, 45.00, '三线城市部门经理早餐标准'),
('ms-3-2-lunch', 3, '部门经理', 'lunch', 120.00, 60.00, '三线城市部门经理午餐标准'),
('ms-3-2-dinner', 3, '部门经理', 'dinner', 180.00, 90.00, '三线城市部门经理晚餐标准'),

('ms-3-3-breakfast', 3, '高管', 'breakfast', 120.00, 60.00, '三线城市高管早餐标准'),
('ms-3-3-lunch', 3, '高管', 'lunch', 180.00, 90.00, '三线城市高管午餐标准'),
('ms-3-3-dinner', 3, '高管', 'dinner', 240.00, 120.00, '三线城市高管晚餐标准');
