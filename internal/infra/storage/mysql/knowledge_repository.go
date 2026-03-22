package mysql

import (
	"reimbursement-audit/internal/domain/knowledge"
	"reimbursement-audit/internal/pkg/logger"

	"gorm.io/gorm"
)

type KnowledgeRepository struct {
	client *Client
	logger logger.Logger
}

func NewKnowledgeRepository(client *Client, logger logger.Logger) knowledge.KnowledgeRepository {
	return &KnowledgeRepository{
		client: client,
		logger: logger,
	}
}

func (r *KnowledgeRepository) getDB() *gorm.DB {
	return r.client.GetDB()
}

func (r *KnowledgeRepository) Create(file *knowledge.KnowledgeFile) error {
	return r.getDB().Create(file).Error
}

func (r *KnowledgeRepository) GetByID(id string) (*knowledge.KnowledgeFile, error) {
	var file knowledge.KnowledgeFile
	err := r.getDB().Where("id = ?", id).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *KnowledgeRepository) GetAll(filter *knowledge.KnowledgeFileFilter) ([]*knowledge.KnowledgeFile, int64, error) {
	var files []*knowledge.KnowledgeFile
	var total int64

	query := r.getDB().Model(&knowledge.KnowledgeFile{})

	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	query.Count(&total)

	offset := (filter.Page - 1) * filter.PageSize
	err := query.Offset(offset).Limit(filter.PageSize).Order("created_at DESC").Find(&files).Error
	return files, total, err
}

func (r *KnowledgeRepository) Update(file *knowledge.KnowledgeFile) error {
	return r.getDB().Save(file).Error
}

func (r *KnowledgeRepository) Delete(id string) error {
	return r.getDB().Where("id = ?", id).Delete(&knowledge.KnowledgeFile{}).Error
}

func (r *KnowledgeRepository) IncrementDownloadCount(id string) error {
	return r.getDB().Model(&knowledge.KnowledgeFile{}).Where("id = ?", id).
		UpdateColumn("download_count", gorm.Expr("download_count + 1")).Error
}

func (r *KnowledgeRepository) GetByUploader(userID string, filter *knowledge.KnowledgeFileFilter) ([]*knowledge.KnowledgeFile, int64, error) {
	var files []*knowledge.KnowledgeFile
	var total int64

	query := r.getDB().Model(&knowledge.KnowledgeFile{}).Where("uploaded_by = ?", userID)

	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	query.Count(&total)

	offset := (filter.Page - 1) * filter.PageSize
	err := query.Offset(offset).Limit(filter.PageSize).Order("created_at DESC").Find(&files).Error
	return files, total, err
}
