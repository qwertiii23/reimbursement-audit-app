# RAG规则系统架构设计

## 1. 系统概述

RAG（检索增强生成）规则系统通过结合向量检索和生成式AI，实现智能规则推荐、自然语言规则生成、规则相似性检测等功能，提升规则引擎的智能化水平。

## 2. 核心功能

### 2.1 智能规则推荐
基于历史数据和当前场景，自动推荐最合适的规则配置

### 2.2 自然语言规则生成
用户用自然语言描述规则，系统自动生成规则配置

### 2.3 规则相似性检测
检测重复或冲突的规则，避免规则冗余

### 2.4 规则解释和法规依据
为规则触发提供详细的解释和相关法规

### 2.5 智能规则优化
基于执行效果，自动优化规则参数

## 3. 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        用户界面层                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ 规则管理界面  │  │ 规则推荐界面  │  │ 自然语言输入  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                        API层                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ 规则推荐API   │  │ 规则生成API   │  │ 规则检测API   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      RAG服务层                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ 向量检索服务  │  │ 生成式AI服务  │  │ 规则分析服务  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      数据层                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ 向量数据库    │  │ 规则数据库    │  │ 知识库       │      │
│  │ (Milvus/Pinecone)│ (MySQL)       │  │ (法规/案例)  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

## 4. 核心组件设计

### 4.1 向量化服务

```go
package rag

import (
    "context"
    "encoding/json"
)

// EmbeddingService 向量化服务
type EmbeddingService struct {
    embeddingClient EmbeddingClient
}

// RuleEmbedding 规则向量
type RuleEmbedding struct {
    RuleID      string    `json:"rule_id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Vector      []float64 `json:"vector"`
    Category    string    `json:"category"`
    CreatedAt   time.Time `json:"created_at"`
}

// InvoiceEmbedding 发票向量
type InvoiceEmbedding struct {
    InvoiceID   string    `json:"invoice_id"`
    InvoiceData string    `json:"invoice_data"` // JSON字符串
    Vector      []float64 `json:"vector"`
    Category    string    `json:"category"`
    CreatedAt   time.Time `json:"created_at"`
}

// EmbedRule 将规则向量化
func (s *EmbeddingService) EmbedRule(rule *Rule) (*RuleEmbedding, error) {
    // 构建规则文本
    text := s.buildRuleText(rule)
    
    // 调用embedding API
    vector, err := s.embeddingClient.Embed(text)
    if err != nil {
        return nil, err
    }
    
    return &RuleEmbedding{
        RuleID:      rule.ID,
        Name:        rule.Name,
        Description: rule.Description,
        Vector:      vector,
        Category:    rule.Category,
        CreatedAt:   time.Now(),
    }, nil
}

// EmbedInvoice 将发票数据向量化
func (s *EmbeddingService) EmbedInvoice(invoice *Invoice) (*InvoiceEmbedding, error) {
    // 将发票数据转换为JSON字符串
    data, err := json.Marshal(invoice)
    if err != nil {
        return nil, err
    }
    
    // 调用embedding API
    vector, err := s.embeddingClient.Embed(string(data))
    if err != nil {
        return nil, err
    }
    
    return &InvoiceEmbedding{
        InvoiceID:   invoice.ID,
        InvoiceData: string(data),
        Vector:      vector,
        Category:    invoice.Category,
        CreatedAt:   time.Now(),
    }, nil
}

// buildRuleText 构建规则文本
func (s *EmbeddingService) buildRuleText(rule *Rule) string {
    return fmt.Sprintf("规则名称：%s\n规则描述：%s\n规则类别：%s\n优先级：%d",
        rule.Name, rule.Description, rule.Category, rule.Priority)
}
```

### 4.2 向量数据库服务

```go
package rag

import (
    "context"
)

// VectorDatabase 向量数据库接口
type VectorDatabase interface {
    InsertRuleEmbedding(ctx context.Context, embedding *RuleEmbedding) error
    InsertInvoiceEmbedding(ctx context.Context, embedding *InvoiceEmbedding) error
    SearchSimilarRules(ctx context.Context, vector []float64, topK int) ([]*RuleEmbedding, error)
    SearchSimilarInvoices(ctx context.Context, vector []float64, topK int) ([]*InvoiceEmbedding, error)
    DeleteRuleEmbedding(ctx context.Context, ruleID string) error
}

// MilvusVectorDatabase Milvus向量数据库实现
type MilvusVectorDatabase struct {
    client MilvusClient
}

func NewMilvusVectorDatabase(config *MilvusConfig) *MilvusVectorDatabase {
    return &MilvusVectorDatabase{
        client: NewMilvusClient(config),
    }
}

// InsertRuleEmbedding 插入规则向量
func (db *MilvusVectorDatabase) InsertRuleEmbedding(ctx context.Context, embedding *RuleEmbedding) error {
    return db.client.Insert(ctx, "rule_embeddings", []interface{}{embedding})
}

// SearchSimilarRules 搜索相似规则
func (db *MilvusVectorDatabase) SearchSimilarRules(ctx context.Context, vector []float64, topK int) ([]*RuleEmbedding, error) {
    results, err := db.client.Search(ctx, "rule_embeddings", vector, topK)
    if err != nil {
        return nil, err
    }
    
    embeddings := make([]*RuleEmbedding, 0, len(results))
    for _, result := range results {
        embeddings = append(embeddings, result.(*RuleEmbedding))
    }
    
    return embeddings, nil
}
```

### 4.3 规则推荐服务

```go
package rag

import (
    "context"
)

// RuleRecommendationService 规则推荐服务
type RuleRecommendationService struct {
    embeddingService *EmbeddingService
    vectorDB         VectorDatabase
    llmService       LLMService
}

// RuleRecommendation 规则推荐
type RuleRecommendation struct {
    RuleID      string  `json:"rule_id"`
    RuleName    string  `json:"rule_name"`
    Description string  `json:"description"`
    Similarity  float64 `json:"similarity"`
    Reason      string  `json:"reason"`
}

// RecommendRules 推荐规则
func (s *RuleRecommendationService) RecommendRules(ctx context.Context, invoice *Invoice, topK int) ([]*RuleRecommendation, error) {
    // 1. 将发票数据向量化
    invoiceEmbedding, err := s.embeddingService.EmbedInvoice(invoice)
    if err != nil {
        return nil, err
    }
    
    // 2. 在向量数据库中搜索相似规则
    similarRules, err := s.vectorDB.SearchSimilarRules(ctx, invoiceEmbedding.Vector, topK)
    if err != nil {
        return nil, err
    }
    
    // 3. 生成推荐理由
    recommendations := make([]*RuleRecommendation, 0, len(similarRules))
    for _, rule := range similarRules {
        reason, err := s.generateRecommendationReason(invoice, rule)
        if err != nil {
            continue
        }
        
        recommendations = append(recommendations, &RuleRecommendation{
            RuleID:      rule.RuleID,
            RuleName:    rule.Name,
            Description: rule.Description,
            Similarity:  calculateSimilarity(invoiceEmbedding.Vector, rule.Vector),
            Reason:      reason,
        })
    }
    
    return recommendations, nil
}

// generateRecommendationReason 生成推荐理由
func (s *RuleRecommendationService) generateRecommendationReason(invoice *Invoice, rule *RuleEmbedding) (string, error) {
    prompt := fmt.Sprintf(`
    基于以下发票数据和规则，生成推荐理由：
    
    发票数据：
    - 金额：%.2f
    - 类型：%s
    - 日期：%s
    - 商户：%s
    
    规则信息：
    - 名称：%s
    - 描述：%s
    
    请生成一个简洁的推荐理由，说明为什么这个规则适用于该发票。
    `, invoice.Amount, invoice.InvoiceType, invoice.InvoiceDate, invoice.MerchantName, rule.Name, rule.Description)
    
    return s.llmService.Generate(prompt)
}

// calculateSimilarity 计算相似度
func calculateSimilarity(vec1, vec2 []float64) float64 {
    // 使用余弦相似度
    dotProduct := 0.0
    norm1 := 0.0
    norm2 := 0.0
    
    for i := 0; i < len(vec1); i++ {
        dotProduct += vec1[i] * vec2[i]
        norm1 += vec1[i] * vec1[i]
        norm2 += vec2[i] * vec2[i]
    }
    
    return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}
```

### 4.4 自然语言规则生成服务

```go
package rag

import (
    "context"
    "encoding/json"
)

// NLRuleGenerationService 自然语言规则生成服务
type NLRuleGenerationService struct {
    llmService       LLMService
    embeddingService *EmbeddingService
    vectorDB         VectorDatabase
}

// GeneratedRule 生成的规则
type GeneratedRule struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Priority    int                    `json:"priority"`
    Feature     *GeneratedFeature      `json:"feature"`
    Condition   *GeneratedCondition    `json:"condition"`
    Decision    *GeneratedDecision     `json:"decision"`
}

// GeneratedFeature 生成的特征
type GeneratedFeature struct {
    Name           string                 `json:"name"`
    Code           string                 `json:"code"`
    FunctionName   string                 `json:"function_name"`
    FunctionConfig map[string]interface{} `json:"function_config"`
}

// GeneratedCondition 生成的条件
type GeneratedCondition struct {
    Operator string `json:"operator"`
    Value    string `json:"value"`
}

// GeneratedDecision 生成的决策
type GeneratedDecision struct {
    Type   string `json:"type"`
    Reason string `json:"reason"`
}

// GenerateRuleFromText 从自然语言生成规则
func (s *NLRuleGenerationService) GenerateRuleFromText(ctx context.Context, description string) (*GeneratedRule, error) {
    // 1. 检索相似规则作为参考
    similarRules, err := s.retrieveSimilarRules(description)
    if err != nil {
        return nil, err
    }
    
    // 2. 构建prompt
    prompt := s.buildGenerationPrompt(description, similarRules)
    
    // 3. 调用LLM生成规则
    response, err := s.llmService.Generate(prompt)
    if err != nil {
        return nil, err
    }
    
    // 4. 解析生成的规则
    var rule GeneratedRule
    if err := json.Unmarshal([]byte(response), &rule); err != nil {
        return nil, err
    }
    
    return &rule, nil
}

// retrieveSimilarRules 检索相似规则
func (s *NLRuleGenerationService) retrieveSimilarRules(description string) ([]*RuleEmbedding, error) {
    // 将描述向量化
    vector, err := s.embeddingService.embeddingClient.Embed(description)
    if err != nil {
        return nil, err
    }
    
    // 搜索相似规则
    return s.vectorDB.SearchSimilarRules(context.Background(), vector, 3)
}

// buildGenerationPrompt 构建生成prompt
func (s *NLRuleGenerationService) buildGenerationPrompt(description string, similarRules []*RuleEmbedding) string {
    examples := ""
    for i, rule := range similarRules {
        examples += fmt.Sprintf(`
        示例%d：
        - 名称：%s
        - 描述：%s
        `, i+1, rule.Name, rule.Description)
    }
    
    return fmt.Sprintf(`
    你是一个规则生成专家。请根据用户的描述，生成一个规则配置。
    
    用户描述：%s
    
    参考示例：%s
    
    请生成一个JSON格式的规则配置，包含以下字段：
    - name: 规则名称
    - description: 规则描述
    - priority: 规则优先级（1-100）
    - feature: 特征配置
      - name: 特征名称
      - code: 特征代码
      - function_name: 特征函数名称
      - function_config: 特征函数配置
    - condition: 条件配置
      - operator: 操作符（eq, gt, lt, gte, lte, contains等）
      - value: 条件值
    - decision: 决策配置
      - type: 决策类型（approve, reject, review）
      - reason: 决策原因
    
    只返回JSON，不要包含其他内容。
    `, description, examples)
}
```

### 4.5 规则相似性检测服务

```go
package rag

import (
    "context"
)

// RuleSimilarityService 规则相似性检测服务
type RuleSimilarityService struct {
    embeddingService *EmbeddingService
    vectorDB         VectorDatabase
    threshold        float64
}

// SimilarRule 相似规则
type SimilarRule struct {
    RuleID      string  `json:"rule_id"`
    RuleName    string  `json:"rule_name"`
    Description string  `json:"description"`
    Similarity  float64 `json:"similarity"`
    Conflict    bool    `json:"conflict"`
    Reason      string  `json:"reason"`
}

// DetectSimilarRules 检测相似规则
func (s *RuleSimilarityService) DetectSimilarRules(ctx context.Context, rule *Rule) ([]*SimilarRule, error) {
    // 1. 将规则向量化
    ruleEmbedding, err := s.embeddingService.EmbedRule(rule)
    if err != nil {
        return nil, err
    }
    
    // 2. 搜索相似规则
    similarRules, err := s.vectorDB.SearchSimilarRules(ctx, ruleEmbedding.Vector, 10)
    if err != nil {
        return nil, err
    }
    
    // 3. 过滤相似度低于阈值的规则
    filteredRules := make([]*SimilarRule, 0)
    for _, similarRule := range similarRules {
        similarity := calculateSimilarity(ruleEmbedding.Vector, similarRule.Vector)
        
        if similarity >= s.threshold {
            conflict := s.analyzeConflict(rule, similarRule)
            reason := s.generateSimilarityReason(rule, similarRule, similarity)
            
            filteredRules = append(filteredRules, &SimilarRule{
                RuleID:      similarRule.RuleID,
                RuleName:    similarRule.Name,
                Description: similarRule.Description,
                Similarity:  similarity,
                Conflict:    conflict,
                Reason:      reason,
            })
        }
    }
    
    return filteredRules, nil
}

// analyzeConflict 分析规则冲突
func (s *RuleSimilarityService) analyzeConflict(rule1 *Rule, rule2 *RuleEmbedding) bool {
    // 检查规则条件是否冲突
    // 例如：一个规则是"金额>1000拒绝"，另一个是"金额>1000审批"
    
    // 这里简化处理，实际需要更复杂的冲突检测逻辑
    if rule1.Priority == rule2.Priority {
        return true
    }
    
    return false
}

// generateSimilarityReason 生成相似性原因
func (s *RuleSimilarityService) generateSimilarityReason(rule1 *Rule, rule2 *RuleEmbedding, similarity float64) string {
    return fmt.Sprintf("规则相似度%.2f，可能存在重复或冲突", similarity)
}
```

## 5. 数据模型

### 5.1 规则向量表

```sql
CREATE TABLE rule_embeddings (
    id VARCHAR(36) PRIMARY KEY,
    rule_id VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    vector JSON NOT NULL,
    category VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_rule_id (rule_id),
    INDEX idx_category (category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='规则向量表';
```

### 5.2 发票向量表

```sql
CREATE TABLE invoice_embeddings (
    id VARCHAR(36) PRIMARY KEY,
    invoice_id VARCHAR(36) NOT NULL,
    invoice_data JSON NOT NULL,
    vector JSON NOT NULL,
    category VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_invoice_id (invoice_id),
    INDEX idx_category (category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='发票向量表';
```

### 5.3 规则推荐记录表

```sql
CREATE TABLE rule_recommendations (
    id VARCHAR(36) PRIMARY KEY,
    invoice_id VARCHAR(36) NOT NULL,
    rule_id VARCHAR(36) NOT NULL,
    similarity FLOAT NOT NULL,
    reason TEXT,
    adopted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_invoice_id (invoice_id),
    INDEX idx_rule_id (rule_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='规则推荐记录表';
```

## 6. API设计

### 6.1 规则推荐API

```go
// POST /api/v1/rag/recommend-rules
type RecommendRulesRequest struct {
    InvoiceID string `json:"invoice_id"`
    TopK      int    `json:"top_k"`
}

type RecommendRulesResponse struct {
    Recommendations []*RuleRecommendation `json:"recommendations"`
}
```

### 6.2 自然语言规则生成API

```go
// POST /api/v1/rag/generate-rule
type GenerateRuleRequest struct {
    Description string `json:"description"`
}

type GenerateRuleResponse struct {
    Rule *GeneratedRule `json:"rule"`
}
```

### 6.3 规则相似性检测API

```go
// POST /api/v1/rag/detect-similar-rules
type DetectSimilarRulesRequest struct {
    Rule *Rule `json:"rule"`
}

type DetectSimilarRulesResponse struct {
    SimilarRules []*SimilarRule `json:"similar_rules"`
}
```

## 7. 实现步骤

### 阶段1：基础设施搭建
1. 选择并部署向量数据库（Milvus/Pinecone）
2. 集成Embedding服务（OpenAI/Cohere）
3. 集成LLM服务（OpenAI GPT-4/Claude）
4. 创建向量表结构

### 阶段2：向量化服务
1. 实现规则向量化
2. 实现发票数据向量化
3. 批量向量化现有规则
4. 建立向量索引

### 阶段3：规则推荐
1. 实现相似规则检索
2. 实现推荐理由生成
3. 实现推荐API
4. 前端界面集成

### 阶段4：自然语言规则生成
1. 实现prompt工程
2. 实现规则生成逻辑
3. 实现生成API
4. 前端界面集成

### 阶段5：规则相似性检测
1. 实现相似度计算
2. 实现冲突检测
3. 实现检测API
4. 前端界面集成

### 阶段6：优化和扩展
1. 性能优化
2. 准确率提升
3. 更多功能扩展
4. 用户反馈收集

## 8. 技术栈

### 后端
- Go 1.21+
- MySQL 8.0+
- Milvus 2.3+ 或 Pinecone
- OpenAI API 或 Cohere API

### 前端
- Vue 3
- TypeScript
- Element Plus

### AI服务
- OpenAI GPT-4
- OpenAI Embeddings
- 或 Cohere API

## 9. 预期效果

### 9.1 智能规则推荐
- 准确率：>85%
- 响应时间：<500ms
- 推荐数量：3-5条

### 9.2 自然语言规则生成
- 成功率：>80%
- 生成时间：<2s
- 规则可用性：>75%

### 9.3 规则相似性检测
- 检测准确率：>90%
- 误报率：<10%
- 响应时间：<300ms

## 10. 风险和挑战

### 10.1 技术风险
- 向量数据库性能
- Embedding质量
- LLM生成稳定性

### 10.2 业务风险
- 规则推荐不准确
- 生成规则不可用
- 相似性检测误报

### 10.3 应对措施
- 充分的测试和验证
- 用户反馈机制
- 持续优化和迭代
