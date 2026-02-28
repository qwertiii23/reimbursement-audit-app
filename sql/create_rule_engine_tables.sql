-- 创建规则引擎相关表
-- 用于存储规则引擎的特征、条件、决策等

CREATE TABLE IF NOT EXISTS features (
    id VARCHAR(36) PRIMARY KEY COMMENT '特征ID',
    name VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '特征名称',
    code VARCHAR(50) NOT NULL UNIQUE COMMENT '特征编码',
    description TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '特征描述',
    type VARCHAR(20) NOT NULL COMMENT '特征类型(string/number/boolean/date)',
    value_type VARCHAR(20) NOT NULL COMMENT '值类型(single/multiple/range)',
    category VARCHAR(50) COMMENT '特征分类',
    enabled BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_code (code),
    INDEX idx_category (category),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='特征表';

CREATE TABLE IF NOT EXISTS feature_values (
    id VARCHAR(36) PRIMARY KEY COMMENT '特征值ID',
    feature_id VARCHAR(36) NOT NULL COMMENT '特征ID',
    value VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '特征值',
    label VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '显示标签',
    sort_order INT DEFAULT 0 COMMENT '排序',
    enabled BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (feature_id) REFERENCES features(id) ON DELETE CASCADE,
    INDEX idx_feature_id (feature_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='特征值表';

CREATE TABLE IF NOT EXISTS rule_engine_rules (
    id VARCHAR(36) PRIMARY KEY COMMENT '规则ID',
    name VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '规则名称',
    description TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '规则描述',
    priority INT NOT NULL DEFAULT 0 COMMENT '优先级(数字越大优先级越高)',
    enabled BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_priority (priority),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='规则引擎规则表';

CREATE TABLE IF NOT EXISTS conditions (
    id VARCHAR(36) PRIMARY KEY COMMENT '条件ID',
    rule_id VARCHAR(36) NOT NULL COMMENT '规则ID',
    feature_id VARCHAR(36) NOT NULL COMMENT '特征ID',
    operator VARCHAR(20) NOT NULL COMMENT '操作符(eq/ne/gt/lt/ge/le/in/not_in/contains)',
    value VARCHAR(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '比较值',
    value_list JSON COMMENT '值列表',
    logic_op VARCHAR(10) DEFAULT 'and' COMMENT '逻辑操作符(and/or)',
    sort_order INT DEFAULT 0 COMMENT '排序',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (rule_id) REFERENCES rule_engine_rules(id) ON DELETE CASCADE,
    FOREIGN KEY (feature_id) REFERENCES features(id) ON DELETE CASCADE,
    INDEX idx_rule_id (rule_id),
    INDEX idx_feature_id (feature_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='条件表';

CREATE TABLE IF NOT EXISTS decisions (
    id VARCHAR(36) PRIMARY KEY COMMENT '决策ID',
    rule_id VARCHAR(36) NOT NULL COMMENT '规则ID',
    type VARCHAR(20) NOT NULL COMMENT '决策类型(approve/reject/review)',
    reason TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '决策原因',
    action VARCHAR(50) COMMENT '执行动作',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (rule_id) REFERENCES rule_engine_rules(id) ON DELETE CASCADE,
    UNIQUE KEY uk_rule_id (rule_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='决策表';
