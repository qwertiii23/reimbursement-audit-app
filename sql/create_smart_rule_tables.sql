-- 城市等级表
CREATE TABLE IF NOT EXISTS city_tier (
    id VARCHAR(36) PRIMARY KEY,
    city_name VARCHAR(50) NOT NULL UNIQUE,
    tier TINYINT NOT NULL COMMENT '城市等级：1-一线城市，2-二线城市，3-三线城市',
    province VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_city_name (city_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='城市等级表';

-- 初始化城市等级数据
INSERT INTO city_tier (id, city_name, tier, province) VALUES
('1', '北京', 1, '北京'),
('2', '上海', 1, '上海'),
('3', '广州', 1, '广东'),
('4', '深圳', 1, '广东'),
('5', '杭州', 2, '浙江'),
('6', '南京', 2, '江苏'),
('7', '武汉', 2, '湖北'),
('8', '成都', 2, '四川'),
('9', '重庆', 2, '重庆'),
('10', '西安', 2, '陕西'),
('11', '天津', 2, '天津'),
('12', '苏州', 2, '江苏'),
('13', '郑州', 2, '河南'),
('14', '长沙', 2, '湖南'),
('15', '青岛', 2, '山东')
ON DUPLICATE KEY UPDATE tier=VALUES(tier), province=VALUES(province);

-- 住宿费标准表
CREATE TABLE IF NOT EXISTS accommodation_standard (
    id VARCHAR(36) PRIMARY KEY,
    city_tier TINYINT NOT NULL,
    employee_level VARCHAR(20) NOT NULL COMMENT '员工职级：普通员工、部门经理、高管',
    daily_limit DECIMAL(10,2) NOT NULL COMMENT '每日限额',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_city_level (city_tier, employee_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='住宿费标准表';

-- 初始化住宿费标准数据
INSERT INTO accommodation_standard (id, city_tier, employee_level, daily_limit) VALUES
('1', 1, '普通员工', 500.00),
('2', 1, '部门经理', 800.00),
('3', 1, '高管', 1200.00),
('4', 2, '普通员工', 400.00),
('5', 2, '部门经理', 600.00),
('6', 2, '高管', 1000.00),
('7', 3, '普通员工', 300.00),
('8', 3, '部门经理', 500.00),
('9', 3, '高管', 800.00)
ON DUPLICATE KEY UPDATE daily_limit=VALUES(daily_limit);

-- 城市活动表
CREATE TABLE IF NOT EXISTS city_event (
    id VARCHAR(36) PRIMARY KEY,
    city_name VARCHAR(50) NOT NULL,
    event_name VARCHAR(100) NOT NULL,
    event_type VARCHAR(20) NOT NULL COMMENT '活动类型：展会、会议、赛事等',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    accommodation_adjustment DECIMAL(5,2) DEFAULT 1.20 COMMENT '住宿费调整系数',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_city_date (city_name, start_date, end_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='城市活动表';

-- 初始化活动数据
INSERT INTO city_event (id, city_name, event_name, event_type, start_date, end_date, accommodation_adjustment) VALUES
('1', '上海', '中国国际进口博览会', '展会', '2026-11-05', '2026-11-10', 1.30),
('2', '广州', '广交会', '展会', '2026-04-15', '2026-04-19', 1.30),
('3', '深圳', '高交会', '展会', '2026-11-15', '2026-11-19', 1.30),
('4', '北京', '中国国际服务贸易交易会', '展会', '2026-05-28', '2026-06-01', 1.30)
ON DUPLICATE KEY UPDATE event_name=VALUES(event_name), start_date=VALUES(start_date), end_date=VALUES(end_date), accommodation_adjustment=VALUES(accommodation_adjustment);

-- 节假日表
CREATE TABLE IF NOT EXISTS holiday (
    id VARCHAR(36) PRIMARY KEY,
    holiday_name VARCHAR(50) NOT NULL,
    holiday_date DATE NOT NULL,
    is_adjusted BOOLEAN DEFAULT FALSE COMMENT '是否为调休日',
    accommodation_adjustment DECIMAL(5,2) DEFAULT 1.30 COMMENT '住宿费调整系数',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_holiday_date (holiday_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节假日表';

-- 初始化节假日数据
INSERT INTO holiday (id, holiday_name, holiday_date, is_adjusted, accommodation_adjustment) VALUES
('1', '元旦', '2026-01-01', FALSE, 1.30),
('2', '春节', '2026-01-28', FALSE, 1.30),
('3', '春节', '2026-01-29', FALSE, 1.30),
('4', '春节', '2026-01-30', FALSE, 1.30),
('5', '春节', '2026-01-31', FALSE, 1.30),
('6', '清明节', '2026-04-04', FALSE, 1.30),
('7', '劳动节', '2026-05-01', FALSE, 1.30),
('8', '端午节', '2026-05-31', FALSE, 1.30),
('9', '中秋节', '2026-09-25', FALSE, 1.30),
('10', '国庆节', '2026-10-01', FALSE, 1.30),
('11', '国庆节', '2026-10-02', FALSE, 1.30),
('12', '国庆节', '2026-10-03', FALSE, 1.30),
('13', '国庆节', '2026-10-04', FALSE, 1.30),
('14', '国庆节', '2026-10-05', FALSE, 1.30),
('15', '国庆节', '2026-10-06', FALSE, 1.30),
('16', '国庆节', '2026-10-07', FALSE, 1.30)
ON DUPLICATE KEY UPDATE holiday_name=VALUES(holiday_name), is_adjusted=VALUES(is_adjusted), accommodation_adjustment=VALUES(accommodation_adjustment);

-- 政策知识库表
CREATE TABLE IF NOT EXISTS policy_knowledge (
    id VARCHAR(36) PRIMARY KEY,
    policy_type VARCHAR(50) NOT NULL COMMENT '政策类型：住宿费、交通费、招待费等',
    policy_section VARCHAR(100) NOT NULL COMMENT '政策章节',
    policy_content TEXT NOT NULL COMMENT '政策内容',
    keywords JSON COMMENT '关键词',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_policy_type (policy_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='政策知识库表';

-- 初始化政策知识库数据
INSERT INTO policy_knowledge (id, policy_type, policy_section, policy_content, keywords) VALUES
('1', '住宿费', '一线城市标准', '一线城市（北京、上海、广州、深圳）：普通员工每人每天500元，部门经理每人每天800元，高管每人每天1200元。', '["一线城市", "北京", "上海", "广州", "深圳", "住宿费", "标准"]'),
('2', '住宿费', '二线城市标准', '二线城市（省会城市、计划单列市）：普通员工每人每天400元，部门经理每人每天600元，高管每人每天1000元。', '["二线城市", "省会城市", "住宿费", "标准"]'),
('3', '住宿费', '三线城市标准', '三线城市及其他地区：普通员工每人每天300元，部门经理每人每天500元，高管每人每天800元。', '["三线城市", "住宿费", "标准"]'),
('4', '住宿费', '超标处理', '超标住宿需提前报请部门负责人批准，超出部分由个人承担。', '["超标", "住宿", "审批", "个人承担"]'),
('5', '住宿费', '活动期间调整', '出差期间如遇大型活动（如展会、会议等），住宿费标准可适当上浮，具体幅度由财务部门根据实际情况确定。', '["活动", "展会", "会议", "住宿费", "上浮"]'),
('6', '住宿费', '节假日调整', '出差期间如遇法定节假日，住宿费标准可适当上浮，具体幅度由财务部门根据实际情况确定。', '["节假日", "法定节假日", "住宿费", "上浮"]')
ON DUPLICATE KEY UPDATE policy_content=VALUES(policy_content), keywords=VALUES(keywords);
