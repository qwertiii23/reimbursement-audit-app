// model.go RAG相关模型（向量数据、Prompt）
// 功能点：
// 1. 定义文档数据模型
// 2. 定义向量数据模型
// 3. 定义Prompt数据模型
// 4. 定义检索结果模型
// 5. 定义分析结果模型
// 6. 提供模型转换和验证方法

package rag

import (
	"math"
	"time"
)

// Document 文档模型
type Document struct {
	ID        string            `json:"id"`         // 文档ID
	Title     string            `json:"title"`      // 文档标题
	Content   string            `json:"content"`    // 文档内容
	Type      string            `json:"type"`       // 文档类型
	Source    string            `json:"source"`     // 文档来源
	Path      string            `json:"path"`       // 文档路径
	Size      int64             `json:"size"`       // 文档大小
	Metadata  *DocumentMetadata `json:"metadata"`   // 文档元数据
	Chunks    []*DocumentChunk  `json:"chunks"`     // 文档分片
	Status    string            `json:"status"`     // 状态
	CreatedAt time.Time         `json:"created_at"` // 创建时间
	UpdatedAt time.Time         `json:"updated_at"` // 更新时间
	Version   string            `json:"version"`    // 版本号
	Tags      []string          `json:"tags"`       // 标签
}

// DocumentMetadata 文档元数据模型
type DocumentMetadata struct {
	Author      string    `json:"author"`       // 作者
	CreatedAt   time.Time `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`   // 更新时间
	Category    string    `json:"category"`     // 分类
	Department  string    `json:"department"`   // 部门
	EffectiveAt time.Time `json:"effective_at"` // 生效时间
	ExpiresAt   time.Time `json:"expires_at"`   // 失效时间
	Priority    int       `json:"priority"`     // 优先级
	Language    string    `json:"language"`     // 语言
	Summary     string    `json:"summary"`      // 摘要
	Keywords    []string  `json:"keywords"`     // 关键词
}

// DocumentChunk 文档分片模型
type DocumentChunk struct {
	ID         string    `json:"id"`          // 分片ID
	DocumentID string    `json:"document_id"` // 文档ID
	Content    string    `json:"content"`     // 分片内容
	StartPos   int       `json:"start_pos"`   // 起始位置
	EndPos     int       `json:"end_pos"`     // 结束位置
	Vector     []float64 `json:"vector"`      // 向量表示
	CreatedAt  time.Time `json:"created_at"`  // 创建时间
	UpdatedAt  time.Time `json:"updated_at"`  // 更新时间
}

// Vector 向量模型
type Vector struct {
	ID           string                 `json:"id"`            // 向量ID
	DocumentID   string                 `json:"document_id"`   // 文档ID
	ChunkID      string                 `json:"chunk_id"`      // 分片ID
	ChunkContent string                 `json:"chunk_content"` // 分片内容
	Values       []float64              `json:"values"`        // 向量值
	Dimension    int                    `json:"dimension"`     // 向量维度
	Category     string                 `json:"category"`      // 类别（差旅费/招待费/发票校验）
	Metadata     map[string]interface{} `json:"metadata"`      // 元数据
	CreatedAt    time.Time              `json:"created_at"`    // 创建时间
	UpdatedAt    time.Time              `json:"updated_at"`    // 更新时间
}

// VectorSearchResult 向量搜索结果模型
type VectorSearchResult struct {
	ID         string                 `json:"id"`          // 结果ID
	DocumentID string                 `json:"document_id"` // 文档ID
	ChunkID    string                 `json:"chunk_id"`    // 分片ID
	Content    string                 `json:"content"`     // 内容
	Score      float64                `json:"score"`       // 相似度分数
	Metadata   map[string]interface{} `json:"metadata"`    // 元数据
}

// Prompt Prompt模型
type Prompt struct {
	ID        string                 `json:"id"`         // Prompt ID
	Name      string                 `json:"name"`       // Prompt名称
	Template  string                 `json:"template"`   // Prompt模板
	Content   string                 `json:"content"`    // Prompt内容
	Type      string                 `json:"type"`       // Prompt类型
	Variables map[string]interface{} `json:"variables"`  // 变量
	Tokens    int                    `json:"tokens"`     // Token数量
	CreatedBy string                 `json:"created_by"` // 创建人
	CreatedAt time.Time              `json:"created_at"` // 创建时间
	UpdatedAt time.Time              `json:"updated_at"` // 更新时间
	Version   string                 `json:"version"`    // 版本号
	Tags      []string               `json:"tags"`       // 标签
}

// ConversationMessage 对话消息模型
type ConversationMessage struct {
	Role      string    `json:"role"`      // 角色(system/user/assistant)
	Content   string    `json:"content"`   // 消息内容
	Timestamp time.Time `json:"timestamp"` // 时间戳
}

// RAGResult RAG结果模型
type RAGResult struct {
	Query          string           `json:"query"`           // 查询内容
	Documents      []*Document      `json:"documents"`       // 检索到的文档
	Chunks         []*DocumentChunk `json:"chunks"`          // 检索到的分片
	Prompt         string           `json:"prompt"`          // 构建的Prompt
	Response       *LLMResponse     `json:"response"`        // 大模型响应
	AnalysisResult *AnalysisResult  `json:"analysis_result"` // 分析结果
	ExecutionTime  int64            `json:"execution_time"`  // 执行时间(毫秒)
	CreatedAt      time.Time        `json:"created_at"`      // 创建时间
}

// LLMResponse 大模型响应模型
type LLMResponse struct {
	ID        string    `json:"id"`         // 响应ID
	Content   string    `json:"content"`    // 响应内容
	Model     string    `json:"model"`      // 模型名称
	Tokens    int       `json:"tokens"`     // Token数量
	Cost      float64   `json:"cost"`       // 成本
	Duration  int64     `json:"duration"`   // 响应时间(毫秒)
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// AnalysisResult 分析结果模型
type AnalysisResult struct {
	ID          string                 `json:"id"`          // 分析结果ID
	Query       string                 `json:"query"`       // 查询内容
	Conclusion  string                 `json:"conclusion"`  // 结论
	Reasoning   string                 `json:"reasoning"`   // 推理过程
	Suggestions []string               `json:"suggestions"` // 建议
	Confidence  float64                `json:"confidence"`  // 置信度
	Data        map[string]interface{} `json:"data"`        // 相关数据
	CreatedAt   time.Time              `json:"created_at"`  // 创建时间
}

// DocumentFilter 文档过滤器模型
type DocumentFilter struct {
	Type       string    `json:"type"`       // 文档类型
	Category   string    `json:"category"`   // 分类
	Department string    `json:"department"` // 部门
	Status     string    `json:"status"`     // 状态
	Tags       []string  `json:"tags"`       // 标签
	StartDate  time.Time `json:"start_date"` // 开始日期
	EndDate    time.Time `json:"end_date"`   // 结束日期
	Page       int       `json:"page"`       // 页码
	Size       int       `json:"size"`       // 每页大小
}

// VectorStoreStatistics 向量存储统计模型
type VectorStoreStatistics struct {
	DocumentCount int64     `json:"document_count"` // 文档数量
	ChunkCount    int64     `json:"chunk_count"`    // 分片数量
	VectorCount   int64     `json:"vector_count"`   // 向量数量
	IndexSize     int64     `json:"index_size"`     // 索引大小
	StorageSize   int64     `json:"storage_size"`   // 存储大小
	LastUpdated   time.Time `json:"last_updated"`   // 最后更新时间
}

// IsValid 检查文档是否有效
func (d *Document) IsValid() bool {
	if d == nil {
		return false
	}
	if d.ID == "" || d.Content == "" {
		return false
	}
	return true
}

// IsExpired 检查文档是否过期
func (d *Document) IsExpired() bool {
	if d.Metadata == nil {
		return false
	}
	if d.Metadata.ExpiresAt.IsZero() {
		return false
	}
	return d.Metadata.ExpiresAt.Before(time.Now())
}

// GetChunkCount 获取分片数量
func (d *Document) GetChunkCount() int {
	if d.Chunks == nil {
		return 0
	}
	return len(d.Chunks)
}

// GetTotalTokens 获取总Token数量
func (d *Document) GetTotalTokens() int {
	if d.Chunks == nil {
		return 0
	}
	total := 0
	for _, chunk := range d.Chunks {
		total += len(chunk.Content) / 4
	}
	return total
}

// IsValid 检查向量是否有效
func (v *Vector) IsValid() bool {
	if v == nil {
		return false
	}
	if v.ID == "" || v.DocumentID == "" {
		return false
	}
	if len(v.Values) == 0 {
		return false
	}
	return true
}

// GetDimension 获取向量维度
func (v *Vector) GetDimension() int {
	if v == nil || v.Values == nil {
		return 0
	}
	return len(v.Values)
}

// Normalize 归一化向量
func (v *Vector) Normalize() {
	if v == nil || v.Values == nil {
		return
	}

	norm := 0.0
	for _, val := range v.Values {
		norm += val * val
	}
	norm = math.Sqrt(norm)

	if norm == 0 {
		return
	}

	for i := range v.Values {
		v.Values[i] = v.Values[i] / norm
	}
}

// IsValid 检查Prompt是否有效
func (p *Prompt) IsValid() bool {
	if p == nil {
		return false
	}
	if p.ID == "" || p.Content == "" {
		return false
	}
	return true
}

// EstimateTokens 估算Token数量
func (p *Prompt) EstimateTokens() int {
	if p == nil || p.Content == "" {
		return 0
	}
	return len(p.Content) / 4
}

// IsHighConfidence 检查分析结果是否为高置信度
func (a *AnalysisResult) IsHighConfidence() bool {
	if a == nil {
		return false
	}
	return a.Confidence >= 0.8
}
