package postgres

import (
	"reimbursement-audit/internal/domain/knowledge"
	"reimbursement-audit/internal/pkg/logger"
	"time"

	"gorm.io/gorm"
)

type KnowledgeRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewKnowledgeRepository(db *gorm.DB, logger logger.Logger) *KnowledgeRepository {
	return &KnowledgeRepository{
		db:     db,
		logger: logger,
	}
}

type DocumentModel struct {
	ID           string    `gorm:"primaryKey;column:id"`
	FileName     string    `gorm:"column:file_name;index"`
	FileType     string    `gorm:"column:file_type"`
	Category     string    `gorm:"column:category"`
	ChunkID      string    `gorm:"column:chunk_id;index"`
	ChunkIndex   int       `gorm:"column:chunk_index"`
	ChunkContent string    `gorm:"column:chunk_content"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (DocumentModel) TableName() string {
	return "reimbursement_documents"
}

func (r *KnowledgeRepository) Create(file *knowledge.KnowledgeFile) error {
	doc := &DocumentModel{
		ID:           file.ID,
		FileName:     file.FileName,
		FileType:     file.FileType,
		Category:     file.Category,
		ChunkID:      file.ID,
		ChunkIndex:   0,
		ChunkContent: file.Description,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	return r.db.Create(doc).Error
}

func (r *KnowledgeRepository) GetByID(id string) (*knowledge.KnowledgeFile, error) {
	var doc DocumentModel
	err := r.db.Where("file_name = ?", id).First(&doc).Error
	if err != nil {
		return nil, err
	}

	file := &knowledge.KnowledgeFile{
		ID:            doc.FileName,
		FileName:      doc.FileName,
		FilePath:      doc.FileName,
		FileType:      doc.FileType,
		FileSize:      0,
		Category:      doc.Category,
		Description:   doc.ChunkContent,
		UploadedBy:    "",
		UploaderName:  "系统",
		DownloadCount: 0,
		Status:        "active",
		CreatedAt:     doc.CreatedAt,
		UpdatedAt:     doc.UpdatedAt,
	}

	return file, nil
}

func (r *KnowledgeRepository) GetAll(filter *knowledge.KnowledgeFileFilter) ([]*knowledge.KnowledgeFile, int64, error) {
	var total int64
	countQuery := r.db.Model(&DocumentModel{}).Select("DISTINCT file_name")
	if filter.Category != "" {
		countQuery = countQuery.Where("category = ?", filter.Category)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize

	type FileSummary struct {
		FileName     string
		FileType     string
		Category     string
		ChunkContent string
		CreatedAt    time.Time
		UpdatedAt    time.Time
		ChunkCount   int
	}

	query := r.db.Table("reimbursement_documents").
		Select("file_name, file_type, category, MIN(chunk_content) as chunk_content, MIN(created_at) as created_at, MAX(updated_at) as updated_at, COUNT(*) as chunk_count")

	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}

	var summaries []FileSummary
	err := query.Group("file_name, file_type, category").
		Order("file_name").
		Limit(filter.PageSize).
		Offset(offset).
		Scan(&summaries).Error

	if err != nil {
		return nil, 0, err
	}

	files := make([]*knowledge.KnowledgeFile, len(summaries))
	for i, summary := range summaries {
		files[i] = &knowledge.KnowledgeFile{
			ID:            summary.FileName,
			FileName:      summary.FileName,
			FilePath:      summary.FileName,
			FileType:      summary.FileType,
			FileSize:      int64(summary.ChunkCount * 1000),
			Category:      summary.Category,
			Description:   summary.ChunkContent,
			UploadedBy:    "",
			UploaderName:  "系统",
			DownloadCount: 0,
			Status:        "active",
			CreatedAt:     summary.CreatedAt,
			UpdatedAt:     summary.UpdatedAt,
		}
	}

	return files, total, nil
}

func (r *KnowledgeRepository) Update(file *knowledge.KnowledgeFile) error {
	return r.db.Model(&DocumentModel{}).
		Where("file_name = ?", file.FileName).
		Update("category", file.Category).Error
}

func (r *KnowledgeRepository) Delete(id string) error {
	return r.db.Where("file_name = ?", id).Delete(&DocumentModel{}).Error
}

func (r *KnowledgeRepository) IncrementDownloadCount(id string) error {
	return nil
}

func (r *KnowledgeRepository) GetFileContent(id string) (string, error) {
	var docs []DocumentModel
	err := r.db.Where("file_name = ?", id).
		Order("chunk_index").
		Find(&docs).Error
	if err != nil {
		return "", err
	}

	var content string
	for _, doc := range docs {
		content += doc.ChunkContent + "\n"
	}

	return content, nil
}

func (r *KnowledgeRepository) GetByUploader(userID string, filter *knowledge.KnowledgeFileFilter) ([]*knowledge.KnowledgeFile, int64, error) {
	return r.GetAll(filter)
}
