# 智能规则系统设计方案

## 1. 核心场景

### 场景1：住宿费智能校验
**问题**：员工出差住宿费超过标准，但可能有特殊情况（活动、节假日）需要放宽标准

**传统规则**：
- 一线城市：普通员工500元/天
- 二线城市：普通员工400元/天
- 三线城市：普通员工300元/天

**智能规则**：
- 基础标准：根据城市等级和员工职级确定
- 动态调整：
  - 检查出差地是否有大型活动（展会、会议等）→ 标准上浮20%
  - 检查出差日期是否为节假日 → 标准上浮30%
  - 检查是否有特殊审批 → 根据审批金额确定

**示例**：
- 员工张三，普通员工，出差上海（一线城市）
- 基础标准：500元/天
- 出差期间有国际展会 → 500 × 1.2 = 600元
- 出差日期为国庆节 → 600 × 1.3 = 780元
- 实际住宿：750元 → 通过审核
- 实际住宿：800元 → 需要审批

## 2. 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        报销单据                              │
│  - 员工信息（职级、部门）                                    │
│  - 出差信息（城市、日期、事由）                               │
│  - 费用信息（类型、金额、发票）                               │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    智能规则引擎                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ 基础规则计算  │  │ RAG知识检索   │  │ 外部数据查询   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      决策生成                                │
│  - 基础标准计算                                              │
│  - 动态调整系数                                              │
│  - 最终审核结果                                              │
└─────────────────────────────────────────────────────────────┘
```

## 3. 核心组件

### 3.1 城市等级数据库

```sql
CREATE TABLE city_tier (
    id VARCHAR(36) PRIMARY KEY,
    city_name VARCHAR(50) NOT NULL UNIQUE,
    tier TINYINT NOT NULL COMMENT '城市等级：1-一线城市，2-二线城市，3-三线城市',
    province VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_city_name (city_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='城市等级表';

-- 初始化数据
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
('15', '青岛', 2, '山东');
```

### 3.2 住宿费标准表

```sql
CREATE TABLE accommodation_standard (
    id VARCHAR(36) PRIMARY KEY,
    city_tier TINYINT NOT NULL,
    employee_level VARCHAR(20) NOT NULL COMMENT '员工职级：普通员工、部门经理、高管',
    daily_limit DECIMAL(10,2) NOT NULL COMMENT '每日限额',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_city_level (city_tier, employee_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='住宿费标准表';

-- 初始化数据
INSERT INTO accommodation_standard (id, city_tier, employee_level, daily_limit) VALUES
('1', 1, '普通员工', 500.00),
('2', 1, '部门经理', 800.00),
('3', 1, '高管', 1200.00),
('4', 2, '普通员工', 400.00),
('5', 2, '部门经理', 600.00),
('6', 2, '高管', 1000.00),
('7', 3, '普通员工', 300.00),
('8', 3, '部门经理', 500.00),
('9', 3, '高管', 800.00);
```

### 3.3 活动信息表

```sql
CREATE TABLE city_event (
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

-- 示例数据
INSERT INTO city_event (id, city_name, event_name, event_type, start_date, end_date, accommodation_adjustment) VALUES
('1', '上海', '中国国际进口博览会', '展会', '2026-11-05', '2026-11-10', 1.30),
('2', '广州', '广交会', '展会', '2026-04-15', '2026-04-19', 1.30),
('3', '深圳', '高交会', '展会', '2026-11-15', '2026-11-19', 1.30),
('4', '北京', '中国国际服务贸易交易会', '展会', '2026-05-28', '2026-06-01', 1.30);
```

### 3.4 节假日表

```sql
CREATE TABLE holiday (
    id VARCHAR(36) PRIMARY KEY,
    holiday_name VARCHAR(50) NOT NULL,
    holiday_date DATE NOT NULL,
    is_adjusted BOOLEAN DEFAULT FALSE COMMENT '是否为调休日',
    accommodation_adjustment DECIMAL(5,2) DEFAULT 1.30 COMMENT '住宿费调整系数',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_holiday_date (holiday_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节假日表';

-- 示例数据
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
('16', '国庆节', '2026-10-07', FALSE, 1.30);
```

## 4. 智能特征函数

### 4.1 住宿费智能校验特征函数

```go
package featurefunction

import (
    "context"
    "fmt"
    "reimbursement-audit/internal/pkg/logger"
    "time"
)

type SmartAccommodationValidationFunction struct {
    logger            logger.Logger
    cityTierRepo      CityTierRepository
    accommodationRepo AccommodationStandardRepository
    eventRepo         CityEventRepository
    holidayRepo       HolidayRepository
}

func NewSmartAccommodationValidationFunction(
    cityTierRepo CityTierRepository,
    accommodationRepo AccommodationStandardRepository,
    eventRepo CityEventRepository,
    holidayRepo HolidayRepository,
) *SmartAccommodationValidationFunction {
    return &SmartAccommodationValidationFunction{
        cityTierRepo:      cityTierRepo,
        accommodationRepo: accommodationRepo,
        eventRepo:         eventRepo,
        holidayRepo:       holidayRepo,
    }
}

func (f *SmartAccommodationValidationFunction) SetLogger(logger logger.Logger) {
    f.logger = logger
}

func (f *SmartAccommodationValidationFunction) GetName() string {
    return "smart_accommodation_validation"
}

func (f *SmartAccommodationValidationFunction) GetDescription() string {
    return "智能住宿费校验，根据城市等级、活动、节假日动态调整标准"
}

func (f *SmartAccommodationValidationFunction) GetConfigSchema() *ConfigSchema {
    return &ConfigSchema{
        Fields: []FieldConfig{
            {
                Name:        "enable_event_adjustment",
                Type:        "boolean",
                Label:       "启用活动调整",
                Required:    false,
                Default:     true,
                Description: "是否启用活动期间的住宿费调整",
            },
            {
                Name:        "enable_holiday_adjustment",
                Type:        "boolean",
                Label:       "启用节假日调整",
                Required:    false,
                Default:     true,
                Description: "是否启用节假日的住宿费调整",
            },
        },
    }
}

func (f *SmartAccommodationValidationFunction) Validate(config map[string]interface{}) error {
    return nil
}

func (f *SmartAccommodationValidationFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
    // 1. 获取输入数据
    cityName, _ := input.InvoiceData["city_name"].(string)
    employeeLevel, _ := input.InvoiceData["employee_level"].(string)
    checkInDateStr, _ := input.InvoiceData["check_in_date"].(string)
    checkOutDateStr, _ := input.InvoiceData["check_out_date"].(string)
    actualAmount, _ := input.InvoiceData["amount"].(float64)
    nights, _ := input.InvoiceData["nights"].(float64)

    // 2. 解析日期
    checkInDate, err := time.Parse("2006-01-02", checkInDateStr)
    if err != nil {
        return &FunctionOutput{Error: "无效的入住日期格式"}, nil
    }

    checkOutDate, err := time.Parse("2006-01-02", checkOutDateStr)
    if err != nil {
        return &FunctionOutput{Error: "无效的退房日期格式"}, nil
    }

    // 3. 查询城市等级
    cityTier, err := f.cityTierRepo.GetCityTier(ctx, cityName)
    if err != nil {
        return &FunctionOutput{Error: fmt.Sprintf("查询城市等级失败: %v", err)}, nil
    }

    // 4. 查询基础住宿标准
    baseStandard, err := f.accommodationRepo.GetStandard(ctx, cityTier.Tier, employeeLevel)
    if err != nil {
        return &FunctionOutput{Error: fmt.Sprintf("查询住宿标准失败: %v", err)}, nil
    }

    // 5. 计算动态调整系数
    adjustmentFactors := f.calculateAdjustmentFactors(ctx, cityName, checkInDate, checkOutDate, input.Config)

    // 6. 计算调整后的标准
    adjustedStandard := baseStandard.DailyLimit * adjustmentFactors.TotalFactor

    // 7. 计算总限额
    totalLimit := adjustedStandard * nights

    // 8. 判断是否通过
    isApproved := actualAmount <= totalLimit

    // 9. 生成详细说明
    explanation := f.generateExplanation(
        cityName,
        cityTier.Tier,
        employeeLevel,
        baseStandard.DailyLimit,
        adjustmentFactors,
        adjustedStandard,
        nights,
        totalLimit,
        actualAmount,
        isApproved,
    )

    return &FunctionOutput{
        Value: isApproved,
        Metadata: map[string]interface{}{
            "city_name":          cityName,
            "city_tier":          cityTier.Tier,
            "employee_level":      employeeLevel,
            "base_standard":       baseStandard.DailyLimit,
            "event_factor":       adjustmentFactors.EventFactor,
            "holiday_factor":     adjustmentFactors.HolidayFactor,
            "total_factor":        adjustmentFactors.TotalFactor,
            "adjusted_standard":   adjustedStandard,
            "nights":             nights,
            "total_limit":        totalLimit,
            "actual_amount":      actualAmount,
            "is_approved":        isApproved,
            "explanation":        explanation,
        },
    }, nil
}

type AdjustmentFactors struct {
    EventFactor   float64
    HolidayFactor float64
    TotalFactor   float64
}

func (f *SmartAccommodationValidationFunction) calculateAdjustmentFactors(
    ctx context.Context,
    cityName string,
    checkInDate time.Time,
    checkOutDate time.Time,
    config map[string]interface{},
) *AdjustmentFactors {
    factors := &AdjustmentFactors{
        EventFactor:   1.0,
        HolidayFactor: 1.0,
        TotalFactor:   1.0,
    }

    // 1. 检查活动调整
    if enable, ok := config["enable_event_adjustment"].(bool); ok && enable {
        eventFactor := f.checkEventAdjustment(ctx, cityName, checkInDate, checkOutDate)
        factors.EventFactor = eventFactor
    }

    // 2. 检查节假日调整
    if enable, ok := config["enable_holiday_adjustment"].(bool); ok && enable {
        holidayFactor := f.checkHolidayAdjustment(ctx, checkInDate, checkOutDate)
        factors.HolidayFactor = holidayFactor
    }

    // 3. 计算总调整系数
    factors.TotalFactor = factors.EventFactor * factors.HolidayFactor

    return factors
}

func (f *SmartAccommodationValidationFunction) checkEventAdjustment(
    ctx context.Context,
    cityName string,
    checkInDate time.Time,
    checkOutDate time.Time,
) float64 {
    // 查询出差期间的城市活动
    events, err := f.eventRepo.GetEventsByCityAndDateRange(ctx, cityName, checkInDate, checkOutDate)
    if err != nil {
        f.logger.Error("查询城市活动失败", logger.NewField("error", err.Error()))
        return 1.0
    }

    if len(events) == 0 {
        return 1.0
    }

    // 找出最大的调整系数
    maxFactor := 1.0
    for _, event := range events {
        if event.AccommodationAdjustment > maxFactor {
            maxFactor = event.AccommodationAdjustment
        }
    }

    return maxFactor
}

func (f *SmartAccommodationValidationFunction) checkHolidayAdjustment(
    ctx context.Context,
    checkInDate time.Time,
    checkOutDate time.Time,
) float64 {
    // 查询出差期间的节假日
    holidays, err := f.holidayRepo.GetHolidaysByDateRange(ctx, checkInDate, checkOutDate)
    if err != nil {
        f.logger.Error("查询节假日失败", logger.NewField("error", err.Error()))
        return 1.0
    }

    if len(holidays) == 0 {
        return 1.0
    }

    // 找出最大的调整系数
    maxFactor := 1.0
    for _, holiday := range holidays {
        if holiday.AccommodationAdjustment > maxFactor {
            maxFactor = holiday.AccommodationAdjustment
        }
    }

    return maxFactor
}

func (f *SmartAccommodationValidationFunction) generateExplanation(
    cityName string,
    cityTier int,
    employeeLevel string,
    baseStandard float64,
    factors *AdjustmentFactors,
    adjustedStandard float64,
    nights float64,
    totalLimit float64,
    actualAmount float64,
    isApproved bool,
) string {
    explanation := fmt.Sprintf("【住宿费智能审核】\n")
    explanation += fmt.Sprintf("出差城市：%s（%d线城市）\n", cityName, cityTier)
    explanation += fmt.Sprintf("员工职级：%s\n", employeeLevel)
    explanation += fmt.Sprintf("基础标准：%.2f元/天\n", baseStandard)

    if factors.EventFactor > 1.0 {
        explanation += fmt.Sprintf("活动调整：×%.2f（出差期间有大型活动）\n", factors.EventFactor)
    }

    if factors.HolidayFactor > 1.0 {
        explanation += fmt.Sprintf("节假日调整：×%.2f（出差期间含节假日）\n", factors.HolidayFactor)
    }

    explanation += fmt.Sprintf("调整后标准：%.2f元/天\n", adjustedStandard)
    explanation += fmt.Sprintf("住宿天数：%.0f天\n", nights)
    explanation += fmt.Sprintf("总限额：%.2f元\n", totalLimit)
    explanation += fmt.Sprintf("实际金额：%.2f元\n", actualAmount)

    if isApproved {
        explanation += "✅ 审核通过"
    } else {
        explanation += "❌ 审核不通过，超出限额"
    }

    return explanation
}
```

## 5. RAG知识库设计

### 5.1 知识库表结构

```sql
CREATE TABLE policy_knowledge (
    id VARCHAR(36) PRIMARY KEY,
    policy_type VARCHAR(50) NOT NULL COMMENT '政策类型：住宿费、交通费、招待费等',
    policy_section VARCHAR(100) NOT NULL COMMENT '政策章节',
    policy_content TEXT NOT NULL COMMENT '政策内容',
    keywords JSON COMMENT '关键词',
    embedding JSON COMMENT '向量表示',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_policy_type (policy_type),
    INDEX idx_keywords ((CAST(keywords AS CHAR(255) ARRAY)))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='政策知识库表';
```

### 5.2 知识库数据示例

```sql
INSERT INTO policy_knowledge (id, policy_type, policy_section, policy_content, keywords) VALUES
('1', '住宿费', '一线城市标准', '一线城市（北京、上海、广州、深圳）：普通员工每人每天500元，部门经理每人每天800元，高管每人每天1200元。', '["一线城市", "北京", "上海", "广州", "深圳", "住宿费", "标准"]'),

('2', '住宿费', '二线城市标准', '二线城市（省会城市、计划单列市）：普通员工每人每天400元，部门经理每人每天600元，高管每人每天1000元。', '["二线城市", "省会城市", "住宿费", "标准"]'),

('3', '住宿费', '超标处理', '超标住宿需提前报请部门负责人批准，超出部分由个人承担。', '["超标", "住宿", "审批", "个人承担"]'),

('4', '住宿费', '活动期间调整', '出差期间如遇大型活动（如展会、会议等），住宿费标准可适当上浮，具体幅度由财务部门根据实际情况确定。', '["活动", "展会", "会议", "住宿费", "上浮"]'),

('5', '住宿费', '节假日调整', '出差期间如遇法定节假日，住宿费标准可适当上浮，具体幅度由财务部门根据实际情况确定。', '["节假日", "法定节假日", "住宿费", "上浮"]'),

('6', '交通费', '飞机票标准', '公司高管（总经理、副总经理等）可根据工作需要选择商务舱，其他员工一律乘坐经济舱。', '["飞机票", "商务舱", "经济舱", "高管"]'),

('7', '交通费', '火车票标准', '员工出差应优先选择高铁或动车二等座，特殊情况可选择一等座。', '["火车票", "高铁", "动车", "二等座", "一等座"]'),

('8', '招待费', '招待对象分类', 'A类客人：公司重要客户、合作伙伴、政府官员等。招待标准为餐费350-400元/人/次。', '["招待费", "A类客人", "重要客户", "标准"]'),

('9', '招待费', '审批权限', '单次招待费超过2000元的，需报请分管领导批准。单次招待费超过5000元的，需报请总经理批准。', '["招待费", "审批", "2000元", "5000元"]'),

('10', '办公费', '办公用品标准', '单次办公用品采购超过1000元的，需行政部负责人批准。单次办公用品采购超过5000元的，需分管领导批准。', '["办公用品", "1000元", "5000元", "审批"]');
```

## 6. 外部数据查询服务

### 6.1 活动查询API

```go
package external

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type EventQueryService struct {
    httpClient *http.Client
    apiKey     string
}

type CityEvent struct {
    ID                       string    `json:"id"`
    CityName                 string    `json:"city_name"`
    EventName                string    `json:"event_name"`
    EventType                string    `json:"event_type"`
    StartDate                time.Time `json:"start_date"`
    EndDate                  time.Time `json:"end_date"`
    AccommodationAdjustment float64   `json:"accommodation_adjustment"`
}

func NewEventQueryService(apiKey string) *EventQueryService {
    return &EventQueryService{
        httpClient: &http.Client{Timeout: 10 * time.Second},
        apiKey:     apiKey,
    }
}

func (s *EventQueryService) QueryEventsByCityAndDateRange(
    ctx context.Context,
    cityName string,
    startDate time.Time,
    endDate time.Time,
) ([]*CityEvent, error) {
    url := fmt.Sprintf("https://api.events.com/v1/events?city=%s&start=%s&end=%s&api_key=%s",
        cityName,
        startDate.Format("2006-01-02"),
        endDate.Format("2006-01-02"),
        s.apiKey,
    )

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := s.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var events []*CityEvent
    if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
        return nil, err
    }

    return events, nil
}
```

### 6.2 节假日查询API

```go
package external

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type HolidayQueryService struct {
    httpClient *http.Client
    apiKey     string
}

type Holiday struct {
    ID                       string    `json:"id"`
    HolidayName              string    `json:"holiday_name"`
    HolidayDate              time.Time `json:"holiday_date"`
    IsAdjusted              bool      `json:"is_adjusted"`
    AccommodationAdjustment float64   `json:"accommodation_adjustment"`
}

func NewHolidayQueryService(apiKey string) *HolidayQueryService {
    return &HolidayQueryService{
        httpClient: &http.Client{Timeout: 10 * time.Second},
        apiKey:     apiKey,
    }
}

func (s *HolidayQueryService) QueryHolidaysByDateRange(
    ctx context.Context,
    startDate time.Time,
    endDate time.Time,
) ([]*Holiday, error) {
    url := fmt.Sprintf("https://api.holidays.com/v1/holidays?start=%s&end=%s&api_key=%s",
        startDate.Format("2006-01-02"),
        endDate.Format("2006-01-02"),
        s.apiKey,
    )

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := s.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var holidays []*Holiday
    if err := json.NewDecoder(resp.Body).Decode(&holidays); err != nil {
        return nil, err
    }

    return holidays, nil
}
```

## 7. 智能规则SQL脚本

```sql
-- 插入智能住宿费校验规则

-- 插入特征
INSERT INTO features (id, name, code, type, value_type, category, function_name, function_config, description)
VALUES (
    'feat-smart-accommodation',
    '智能住宿费校验',
    'smart_accommodation_validation',
    'boolean',
    'boolean',
    'validation',
    'smart_accommodation_validation',
    '{"enable_event_adjustment": true, "enable_holiday_adjustment": true}',
    '智能住宿费校验，根据城市等级、活动、节假日动态调整标准'
)
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    type = VALUES(type),
    value_type = VALUES(value_type),
    category = VALUES(category),
    function_name = VALUES(function_name),
    function_config = VALUES(function_config),
    description = VALUES(description);

-- 插入规则
INSERT INTO rule_engine_rules (id, name, description, priority, enabled)
VALUES (
    'rule-smart-accommodation',
    '智能住宿费校验规则',
    '根据城市等级、活动、节假日动态调整住宿费标准，智能审核住宿费报销',
    90,
    true
)
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    description = VALUES(description),
    priority = VALUES(priority),
    enabled = VALUES(enabled);

-- 插入条件
INSERT INTO conditions (id, rule_id, feature_id, operator, value, logic_op, sort_order)
VALUES (
    'cond-smart-accommodation',
    'rule-smart-accommodation',
    'feat-smart-accommodation',
    'eq',
    'false',
    'and',
    1
)
ON DUPLICATE KEY UPDATE
    rule_id = VALUES(rule_id),
    feature_id = VALUES(feature_id),
    operator = VALUES(operator),
    value = VALUES(value),
    logic_op = VALUES(logic_op),
    sort_order = VALUES(sort_order);

-- 插入决策
INSERT INTO decisions (id, rule_id, type, reason, created_at, updated_at)
VALUES (
    'decision-smart-accommodation',
    'rule-smart-accommodation',
    'review',
    '住宿费超出智能调整后的标准，需要人工审核',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON DUPLICATE KEY UPDATE
    rule_id = VALUES(rule_id),
    type = VALUES(type),
    reason = VALUES(reason),
    updated_at = CURRENT_TIMESTAMP;
```

## 8. 测试用例

### 8.1 测试数据

```json
{
  "invoice_data": {
    "city_name": "上海",
    "employee_level": "普通员工",
    "check_in_date": "2026-11-06",
    "check_out_date": "2026-11-08",
    "amount": 1800.00,
    "nights": 2
  }
}
```

### 8.2 预期结果

```
【住宿费智能审核】
出差城市：上海（1线城市）
员工职级：普通员工
基础标准：500.00元/天
活动调整：×1.30（出差期间有大型活动：中国国际进口博览会）
节假日调整：×1.00
调整后标准：650.00元/天
住宿天数：2天
总限额：1300.00元
实际金额：1800.00元
❌ 审核不通过，超出限额
```

## 9. 实施步骤

### 阶段1：数据准备（1天）
1. 创建数据库表结构
2. 初始化城市等级数据
3. 初始化住宿费标准数据
4. 初始化节假日数据

### 阶段2：特征函数开发（2天）
1. 实现住宿费智能校验特征函数
2. 实现城市等级查询
3. 实现活动查询
4. 实现节假日查询

### 阶段3：规则配置（1天）
1. 创建智能规则SQL脚本
2. 执行SQL脚本
3. 验证规则配置

### 阶段4：测试验证（1天）
1. 编写测试用例
2. 执行测试
3. 调整参数

### 阶段5：上线部署（1天）
1. 编译部署
2. 监控运行
3. 收集反馈

## 10. 扩展功能

### 10.1 招待费智能校验
- 根据招待对象类型动态调整标准
- 检查招待对象是否在客户名单中
- 根据招待人数计算总额

### 10.2 交通费智能校验
- 根据出差距离推荐交通方式
- 检查票价是否合理
- 比较不同交通方式的价格

### 10.3 办公费智能校验
- 根据部门预算控制采购金额
- 检查采购物品是否在允许清单中
- 比较不同供应商的价格

### 10.4 培训费智能校验
- 根据培训内容评估费用合理性
- 检查培训机构资质
- 比较市场平均价格
