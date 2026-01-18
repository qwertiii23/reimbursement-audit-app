package knowledge

type KnowledgeRepository interface {
	Create(file *KnowledgeFile) error
	GetByID(id string) (*KnowledgeFile, error)
	GetAll(filter *KnowledgeFileFilter) ([]*KnowledgeFile, int64, error)
	Update(file *KnowledgeFile) error
	Delete(id string) error
	IncrementDownloadCount(id string) error
	GetByUploader(userID string, filter *KnowledgeFileFilter) ([]*KnowledgeFile, int64, error)
}
