// vector_store.go PGVector检索封装
// 功能点：
// 1. 向量数据存储和检索
// 2. 相似度搜索
// 3. 向量索引管理
// 4. 向量数据增删改查
// 5. 批量向量操作
// 6. 向量检索性能优化

package rag

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"math"
	"reimbursement-audit/internal/pkg/logger"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	VectorDimension = 768
)

// VectorData 向量数据类型
type VectorData []float64

// Scan 实现 sql.Scanner 接口
func (v *VectorData) Scan(value interface{}) error {
	if value == nil {
		*v = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("无法扫描向量数据")
	}
	var result []float64
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}
	*v = result
	return nil
}

// Value 实现 driver.Valuer 接口
func (v VectorData) Value() (driver.Value, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}

func (v VectorData) GormDataType() string {
	return "vector(768)"
}

func (v VectorData) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	if v == nil {
		return clause.Expr{
			SQL: "NULL",
		}
	}
	data, err := json.Marshal(v)
	if err != nil {
		return clause.Expr{
			SQL: "NULL",
		}
	}
	return clause.Expr{
		SQL:  "?",
		Vars: []interface{}{string(data)},
	}
}

// DocumentModel 文档模型
type DocumentModel struct {
	ID           string     `gorm:"primaryKey;column:id"`
	FileName     string     `gorm:"column:file_name;index"`
	FileType     string     `gorm:"column:file_type"`
	Category     string     `gorm:"column:category"`
	ChunkID      string     `gorm:"column:chunk_id;index"`
	ChunkIndex   int        `gorm:"column:chunk_index"`
	ChunkContent string     `gorm:"column:chunk_content"`
	Embedding    VectorData `gorm:"column:embedding;type:vector(768)"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

// TableName 指定表名
func (DocumentModel) TableName() string {
	return "reimbursement_documents"
}

// VectorStore 向量存储结构体
type VectorStore struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewVectorStore 创建向量存储实例
func NewVectorStore(dsn string, log logger.Logger) (*VectorStore, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Error("连接数据库失败", logger.NewField("error", err))
		return nil, err
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(&DocumentModel{}); err != nil {
		log.Error("迁移表结构失败", logger.NewField("error", err))
		return nil, err
	}

	return &VectorStore{
		db:     db,
		logger: log,
	}, nil
}

// NewVectorStoreWithDB 使用已有的 GORM DB 实例创建向量存储
func NewVectorStoreWithDB(db *gorm.DB, log logger.Logger) *VectorStore {
	return &VectorStore{
		db:     db,
		logger: log,
	}
}

func (vs *VectorStore) validateVector(vector *Vector) error {
	if vector == nil {
		return errors.New("向量不能为空")
	}
	if vector.ID == "" {
		return errors.New("向量ID不能为空")
	}
	if vector.DocumentID == "" {
		return errors.New("文档ID不能为空")
	}
	if len(vector.Values) != VectorDimension {
		return errors.New("向量维度必须为768维")
	}
	return nil
}

func (vs *VectorStore) retryOperation(operation func() error, maxRetries int) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := operation(); err != nil {
			lastErr = err
			vs.logger.Warn("操作失败，重试中", logger.NewField("retry", i+1), logger.NewField("error", err))
			continue
		}
		return nil
	}
	return lastErr
}

// StoreVector 存储向量
func (vs *VectorStore) StoreVector(ctx context.Context, vector *Vector) error {
	if err := vs.validateVector(vector); err != nil {
		vs.logger.Error("向量校验失败", logger.NewField("vector_id", vector.ID), logger.NewField("error", err))
		return err
	}

	if vector.ChunkContent == "" {
		vs.logger.Error("分片内容不能为空", logger.NewField("vector_id", vector.ID))
		return errors.New("分片内容不能为空")
	}

	operation := func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		doc := &DocumentModel{
			ID:           vector.ID,
			FileName:     vector.DocumentID,
			FileType:     "text",
			Category:     vector.Category,
			ChunkID:      vector.ChunkID,
			ChunkIndex:   0,
			ChunkContent: vector.ChunkContent,
			Embedding:    VectorData(vector.Values),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		result := vs.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"embedding", "chunk_content", "category", "updated_at"}),
		}).Create(doc)

		return result.Error
	}

	if err := vs.retryOperation(operation, 2); err != nil {
		vs.logger.Error("存储向量失败", logger.NewField("vector_id", vector.ID), logger.NewField("error", err))
		return err
	}

	return nil
}

// StoreVectors 批量存储向量
func (vs *VectorStore) StoreVectors(ctx context.Context, vectors []*Vector) error {
	if len(vectors) == 0 {
		return nil
	}

	docs := make([]*DocumentModel, 0, len(vectors))
	for _, vector := range vectors {
		if err := vs.validateVector(vector); err != nil {
			vs.logger.Warn("向量校验失败，跳过", logger.NewField("vector_id", vector.ID), logger.NewField("error", err))
			continue
		}

		if vector.ChunkContent == "" {
			vs.logger.Warn("分片内容为空，跳过", logger.NewField("vector_id", vector.ID))
			continue
		}

		doc := &DocumentModel{
			ID:           vector.ID,
			FileName:     vector.DocumentID,
			FileType:     "text",
			Category:     vector.Category,
			ChunkID:      vector.ChunkID,
			ChunkIndex:   0,
			ChunkContent: vector.ChunkContent,
			Embedding:    VectorData(vector.Values),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		docs = append(docs, doc)
	}

	if len(docs) == 0 {
		return nil
	}

	operation := func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		result := vs.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"embedding", "chunk_content", "category", "updated_at"}),
		}).CreateInBatches(docs, 100)

		return result.Error
	}

	if err := vs.retryOperation(operation, 2); err != nil {
		vs.logger.Error("批量存储向量失败", logger.NewField("count", len(docs)), logger.NewField("error", err))
		return err
	}

	return nil
}

// SearchVector 搜索相似向量
func (vs *VectorStore) SearchVector(ctx context.Context, queryVector []float64, topK int) ([]*VectorSearchResult, error) {
	if len(queryVector) == 0 {
		vs.logger.Error("查询向量不能为空")
		return nil, errors.New("查询向量不能为空")
	}

	if len(queryVector) != VectorDimension {
		vs.logger.Error("查询向量维度必须为768维", logger.NewField("dimension", len(queryVector)))
		return nil, errors.New("查询向量维度必须为768维")
	}

	if topK <= 0 {
		topK = 10
	}

	operation := func() ([]*VectorSearchResult, error) {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		type SearchResult struct {
			ID           string
			FileName     string
			FileType     string
			Category     string
			ChunkID      string
			ChunkIndex   int
			ChunkContent string
			Distance     float64
		}

		var results []SearchResult
		queryVectorJSON, _ := json.Marshal(queryVector)

		err := vs.db.WithContext(ctx).Raw(`
			SELECT id, file_name, file_type, category, chunk_id, chunk_index, chunk_content, 
				   embedding <-> ?::vector AS distance
			FROM reimbursement_documents
			WHERE embedding IS NOT NULL
			ORDER BY distance ASC
			LIMIT ?
		`, string(queryVectorJSON), topK).Scan(&results).Error

		if err != nil {
			return nil, err
		}

		vectorResults := make([]*VectorSearchResult, 0, len(results))
		for _, result := range results {
			vectorResults = append(vectorResults, &VectorSearchResult{
				ID:         result.ID,
				DocumentID: result.FileName,
				ChunkID:    result.ChunkID,
				Content:    result.ChunkContent,
				Score:      1.0 - result.Distance,
				Metadata: map[string]interface{}{
					"category":  result.Category,
					"file_type": result.FileType,
				},
			})
		}

		return vectorResults, nil
	}

	results, err := operation()
	if err != nil {
		vs.logger.Error("查询向量失败", logger.NewField("top_k", topK), logger.NewField("error", err))
		return nil, err
	}

	return results, nil
}

func (vs *VectorStore) SearchVectorByCategory(ctx context.Context, queryVector []float64, category string, topK int) ([]*VectorSearchResult, error) {
	if len(queryVector) == 0 {
		vs.logger.Error("查询向量不能为空")
		return nil, errors.New("查询向量不能为空")
	}

	if len(queryVector) != VectorDimension {
		vs.logger.Error("查询向量维度必须为768维", logger.NewField("dimension", len(queryVector)))
		return nil, errors.New("查询向量维度必须为768维")
	}

	if topK <= 0 {
		topK = 10
	}

	operation := func() ([]*VectorSearchResult, error) {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		type SearchResult struct {
			ID           string
			FileName     string
			FileType     string
			Category     string
			ChunkID      string
			ChunkIndex   int
			ChunkContent string
			Distance     float64
		}

		var results []SearchResult
		queryVectorJSON, _ := json.Marshal(queryVector)

		err := vs.db.WithContext(ctx).Raw(`
			SELECT id, file_name, file_type, category, chunk_id, chunk_index, chunk_content, 
				   embedding <-> ?::vector AS distance
			FROM reimbursement_documents
			WHERE embedding IS NOT NULL AND category = ?
			ORDER BY distance ASC
			LIMIT ?
		`, string(queryVectorJSON), category, topK).Scan(&results).Error

		if err != nil {
			return nil, err
		}

		vectorResults := make([]*VectorSearchResult, 0, len(results))
		for _, result := range results {
			vectorResults = append(vectorResults, &VectorSearchResult{
				ID:         result.ID,
				DocumentID: result.FileName,
				ChunkID:    result.ChunkID,
				Content:    result.ChunkContent,
				Score:      1.0 - result.Distance,
				Metadata: map[string]interface{}{
					"category":  result.Category,
					"file_type": result.FileType,
				},
			})
		}

		return vectorResults, nil
	}

	results, err := operation()
	if err != nil {
		vs.logger.Error("按类别查询向量失败", logger.NewField("category", category), logger.NewField("top_k", topK), logger.NewField("error", err))
		return nil, err
	}

	return results, nil
}

// GetVectorByID 根据ID获取向量
func (vs *VectorStore) GetVectorByID(ctx context.Context, id string) (*Vector, error) {
	if id == "" {
		vs.logger.Error("ID不能为空")
		return nil, errors.New("ID不能为空")
	}

	operation := func() (*Vector, error) {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		var doc DocumentModel
		result := vs.db.WithContext(ctx).
			Where("id = ?", id).
			First(&doc)

		if result.Error != nil {
			return nil, result.Error
		}

		vector := &Vector{
			ID:           doc.ID,
			DocumentID:   doc.FileName,
			ChunkID:      doc.ChunkID,
			ChunkContent: doc.ChunkContent,
			Values:       doc.Embedding,
			Dimension:    len(doc.Embedding),
			Category:     doc.Category,
			Metadata:     map[string]interface{}{},
			CreatedAt:    doc.CreatedAt,
			UpdatedAt:    doc.UpdatedAt,
		}

		return vector, nil
	}

	vector, err := operation()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			vs.logger.Error("向量不存在", logger.NewField("id", id))
			return nil, errors.New("向量不存在")
		}
		vs.logger.Error("查询向量失败", logger.NewField("id", id), logger.NewField("error", err))
		return nil, err
	}

	return vector, nil
}

// UpdateVector 更新向量
func (vs *VectorStore) UpdateVector(ctx context.Context, vector *Vector) error {
	if err := vs.validateVector(vector); err != nil {
		vs.logger.Error("向量校验失败", logger.NewField("vector_id", vector.ID), logger.NewField("error", err))
		return err
	}

	operation := func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		result := vs.db.WithContext(ctx).
			Model(&DocumentModel{}).
			Where("id = ?", vector.ID).
			Updates(map[string]interface{}{
				"embedding":     VectorData(vector.Values),
				"chunk_content": vector.ChunkContent,
				"category":      vector.Category,
				"updated_at":    time.Now(),
			})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("向量不存在")
		}

		return nil
	}

	if err := vs.retryOperation(operation, 2); err != nil {
		vs.logger.Error("更新向量失败", logger.NewField("vector_id", vector.ID), logger.NewField("error", err))
		return err
	}

	return nil
}

// DeleteVector 删除向量
func (vs *VectorStore) DeleteVector(ctx context.Context, id string) error {
	if id == "" {
		vs.logger.Error("ID不能为空")
		return errors.New("ID不能为空")
	}

	operation := func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		result := vs.db.WithContext(ctx).
			Where("id = ?", id).
			Delete(&DocumentModel{})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("向量不存在")
		}

		return nil
	}

	if err := vs.retryOperation(operation, 2); err != nil {
		vs.logger.Error("删除向量失败", logger.NewField("id", id), logger.NewField("error", err))
		return err
	}

	return nil
}

// DeleteVectors 批量删除向量
func (vs *VectorStore) DeleteVectors(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	operation := func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		result := vs.db.WithContext(ctx).
			Where("id IN ?", ids).
			Delete(&DocumentModel{})

		return result.Error
	}

	if err := vs.retryOperation(operation, 2); err != nil {
		vs.logger.Error("批量删除向量失败", logger.NewField("count", len(ids)), logger.NewField("error", err))
		return err
	}

	return nil
}

// DeleteVectorByDocument 根据文档ID删除向量
func (vs *VectorStore) DeleteVectorByDocument(ctx context.Context, documentID string) error {
	if documentID == "" {
		vs.logger.Error("文档ID不能为空")
		return errors.New("文档ID不能为空")
	}

	operation := func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		result := vs.db.WithContext(ctx).
			Where("file_name = ?", documentID).
			Delete(&DocumentModel{})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("文档向量不存在")
		}

		return nil
	}

	if err := vs.retryOperation(operation, 2); err != nil {
		vs.logger.Error("删除文档向量失败", logger.NewField("document_id", documentID), logger.NewField("error", err))
		return err
	}

	return nil
}

// GetVectorsByDocumentID 根据文档ID获取向量列表
func (vs *VectorStore) GetVectorsByDocumentID(ctx context.Context, documentID string) ([]*Vector, error) {
	if documentID == "" {
		vs.logger.Error("文档ID不能为空")
		return nil, errors.New("文档ID不能为空")
	}

	operation := func() ([]*Vector, error) {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		var docs []*DocumentModel
		result := vs.db.WithContext(ctx).
			Where("file_name = ? AND embedding IS NOT NULL", documentID).
			Find(&docs)

		if result.Error != nil {
			return nil, result.Error
		}

		vectors := make([]*Vector, 0, len(docs))
		for _, doc := range docs {
			vector := &Vector{
				ID:           doc.ID,
				DocumentID:   doc.FileName,
				ChunkID:      doc.ChunkID,
				ChunkContent: doc.ChunkContent,
				Values:       doc.Embedding,
				Dimension:    len(doc.Embedding),
				Category:     doc.Category,
				Metadata:     map[string]interface{}{},
				CreatedAt:    doc.CreatedAt,
				UpdatedAt:    doc.UpdatedAt,
			}
			vectors = append(vectors, vector)
		}

		return vectors, nil
	}

	vectors, err := operation()
	if err != nil {
		vs.logger.Error("查询向量失败", logger.NewField("document_id", documentID), logger.NewField("error", err))
		return nil, err
	}

	return vectors, nil
}

// CreateIndex 创建向量索引
func (vs *VectorStore) CreateIndex(ctx context.Context, indexName string, indexType string) error {
	if indexName == "" {
		vs.logger.Error("索引名称不能为空")
		return errors.New("索引名称不能为空")
	}

	operation := func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		query := "CREATE INDEX " + indexName + " ON reimbursement_documents(chunk_content)"
		result := vs.db.WithContext(ctx).Exec(query)

		return result.Error
	}

	if err := vs.retryOperation(operation, 2); err != nil {
		vs.logger.Error("创建索引失败", logger.NewField("index_name", indexName), logger.NewField("error", err))
		return err
	}

	return nil
}

func (vs *VectorStore) CreateVectorIndex(ctx context.Context, indexName string, lists int) error {
	if indexName == "" {
		vs.logger.Error("索引名称不能为空")
		return errors.New("索引名称不能为空")
	}

	if lists <= 0 {
		lists = 100
	}

	operation := func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		query := "CREATE INDEX " + indexName + " ON reimbursement_documents USING ivfflat (embedding vector_cosine_ops) WITH (lists = ?)"
		result := vs.db.WithContext(ctx).Exec(query, lists)

		return result.Error
	}

	if err := vs.retryOperation(operation, 2); err != nil {
		vs.logger.Error("创建向量索引失败", logger.NewField("index_name", indexName), logger.NewField("lists", lists), logger.NewField("error", err))
		return err
	}

	return nil
}

// DropIndex 删除向量索引
func (vs *VectorStore) DropIndex(ctx context.Context, indexName string) error {
	if indexName == "" {
		vs.logger.Error("索引名称不能为空")
		return errors.New("索引名称不能为空")
	}

	operation := func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		query := "DROP INDEX " + indexName
		result := vs.db.WithContext(ctx).Exec(query)

		return result.Error
	}

	if err := vs.retryOperation(operation, 2); err != nil {
		vs.logger.Error("删除索引失败", logger.NewField("index_name", indexName), logger.NewField("error", err))
		return err
	}

	return nil
}

// ListIndexes 列出所有索引
func (vs *VectorStore) ListIndexes(ctx context.Context) ([]string, error) {
	query := `
		SELECT INDEX_NAME 
		FROM INFORMATION_SCHEMA.STATISTICS 
		WHERE TABLE_NAME = 'reimbursement_documents'
		GROUP BY INDEX_NAME
	`

	rows, err := vs.db.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		vs.logger.Error("查询索引失败", logger.NewField("error", err))
		return nil, err
	}
	defer rows.Close()

	var indexes []string
	for rows.Next() {
		var indexName string
		if err := rows.Scan(&indexName); err != nil {
			continue
		}
		indexes = append(indexes, indexName)
	}

	if err := rows.Err(); err != nil {
		vs.logger.Error("遍历结果失败", logger.NewField("error", err))
		return nil, err
	}

	return indexes, nil
}

// OptimizeIndex 优化向量索引
func (vs *VectorStore) OptimizeIndex(ctx context.Context, indexName string) error {
	query := "ANALYZE TABLE reimbursement_documents"
	result := vs.db.WithContext(ctx).Exec(query)

	if result.Error != nil {
		vs.logger.Error("优化索引失败", logger.NewField("index_name", indexName), logger.NewField("error", result.Error))
		return result.Error
	}

	return nil
}

// GetStatistics 获取向量存储统计信息
func (vs *VectorStore) GetStatistics(ctx context.Context) (*VectorStoreStatistics, error) {
	stats := &VectorStoreStatistics{
		LastUpdated: time.Now(),
	}

	var documentCount int64
	result := vs.db.WithContext(ctx).
		Model(&DocumentModel{}).
		Distinct("file_name").
		Count(&documentCount)

	if result.Error != nil {
		vs.logger.Error("查询文档数量失败", logger.NewField("error", result.Error))
		return nil, result.Error
	}
	stats.DocumentCount = documentCount

	var chunkCount int64
	result = vs.db.WithContext(ctx).
		Model(&DocumentModel{}).
		Count(&chunkCount)

	if result.Error != nil {
		vs.logger.Error("查询分片数量失败", logger.NewField("error", result.Error))
		return nil, result.Error
	}
	stats.ChunkCount = chunkCount

	var vectorCount int64
	result = vs.db.WithContext(ctx).
		Model(&DocumentModel{}).
		Where("embedding IS NOT NULL").
		Count(&vectorCount)

	if result.Error != nil {
		vs.logger.Error("查询向量数量失败", logger.NewField("error", result.Error))
		return nil, result.Error
	}
	stats.VectorCount = vectorCount

	return stats, nil
}

// HybridSearch 混合搜索（向量+关键词）
func (vs *VectorStore) HybridSearch(ctx context.Context, queryVector []float64, keywords []string, topK int) ([]*VectorSearchResult, error) {
	vectorResults, err := vs.SearchVector(ctx, queryVector, topK*2)
	if err != nil {
		return nil, err
	}

	if len(keywords) == 0 {
		if len(vectorResults) > topK {
			return vectorResults[:topK], nil
		}
		return vectorResults, nil
	}

	keywordResults, err := vs.KeywordSearch(ctx, keywords, topK*2)
	if err != nil {
		return nil, err
	}

	combined := vs.CombineResults(vectorResults, keywordResults, topK)
	return combined, nil
}

// KeywordSearch 关键词搜索
func (vs *VectorStore) KeywordSearch(ctx context.Context, keywords []string, topK int) ([]*VectorSearchResult, error) {
	if len(keywords) == 0 {
		return nil, nil
	}

	query := vs.db.WithContext(ctx).
		Model(&DocumentModel{}).
		Where("chunk_content LIKE ?", "%"+keywords[0]+"%")

	for i := 1; i < len(keywords); i++ {
		query = query.Or("chunk_content LIKE ?", "%"+keywords[i]+"%")
	}

	var docs []*DocumentModel
	result := query.Limit(topK).Find(&docs)

	if result.Error != nil {
		vs.logger.Error("关键词搜索失败", logger.NewField("keywords", strings.Join(keywords, ",")), logger.NewField("error", result.Error))
		return nil, result.Error
	}

	var results []*VectorSearchResult
	for _, doc := range docs {
		results = append(results, &VectorSearchResult{
			ID:         doc.ID,
			DocumentID: doc.FileName,
			ChunkID:    doc.ChunkID,
			Content:    doc.ChunkContent,
			Score:      0.5,
			Metadata:   map[string]interface{}{},
		})
	}

	return results, nil
}

// CombineResults 合并搜索结果
func (vs *VectorStore) CombineResults(vectorResults, keywordResults []*VectorSearchResult, topK int) []*VectorSearchResult {
	scoreMap := make(map[string]*VectorSearchResult)

	for _, result := range vectorResults {
		if existing, ok := scoreMap[result.ID]; ok {
			existing.Score = (existing.Score + result.Score) / 2
		} else {
			scoreMap[result.ID] = result
		}
	}

	for _, result := range keywordResults {
		if existing, ok := scoreMap[result.ID]; ok {
			existing.Score = (existing.Score + result.Score) / 2
		} else {
			scoreMap[result.ID] = result
		}
	}

	var combined []*VectorSearchResult
	for _, result := range scoreMap {
		combined = append(combined, result)
	}

	for i := 0; i < len(combined)-1; i++ {
		for j := i + 1; j < len(combined); j++ {
			if combined[i].Score < combined[j].Score {
				combined[i], combined[j] = combined[j], combined[i]
			}
		}
	}

	if len(combined) > topK {
		combined = combined[:topK]
	}

	return combined
}

// FilterSearch 过滤搜索
func (vs *VectorStore) FilterSearch(ctx context.Context, queryVector []float64, filters map[string]interface{}, topK int) ([]*VectorSearchResult, error) {
	vectorResults, err := vs.SearchVector(ctx, queryVector, topK*5)
	if err != nil {
		return nil, err
	}

	var filtered []*VectorSearchResult
	for _, result := range vectorResults {
		match := true
		for key, value := range filters {
			if resultVal, ok := result.Metadata[key]; ok {
				if resultVal != value {
					match = false
					break
				}
			}
		}
		if match {
			filtered = append(filtered, result)
		}
	}

	if len(filtered) > topK {
		filtered = filtered[:topK]
	}

	return filtered, nil
}

// CalculateSimilarity 计算向量相似度
func (vs *VectorStore) CalculateSimilarity(vector1, vector2 []float64) float64 {
	if len(vector1) != len(vector2) {
		return 0
	}

	var dotProduct, norm1, norm2 float64
	for i := 0; i < len(vector1); i++ {
		dotProduct += vector1[i] * vector2[i]
		norm1 += vector1[i] * vector1[i]
		norm2 += vector2[i] * vector2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// NormalizeVector 向量归一化
func (vs *VectorStore) NormalizeVector(vector []float64) []float64 {
	if len(vector) == 0 {
		return vector
	}

	norm := 0.0
	for _, v := range vector {
		norm += v * v
	}
	norm = math.Sqrt(norm)

	if norm == 0 {
		return vector
	}

	normalized := make([]float64, len(vector))
	for i, v := range vector {
		normalized[i] = v / norm
	}

	return normalized
}
