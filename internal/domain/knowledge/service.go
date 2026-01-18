package knowledge

type KnowledgeService interface {
	UploadFile(file *KnowledgeFile) error
	GetFileByID(id string) (*KnowledgeFile, error)
	GetAllFiles(filter *KnowledgeFileFilter) ([]*KnowledgeFile, int64, error)
	UpdateFile(file *KnowledgeFile) error
	DeleteFile(id string, userID string) error
	DownloadFile(id string) (*KnowledgeFile, error)
	GetFilesByUploader(userID string, filter *KnowledgeFileFilter) ([]*KnowledgeFile, int64, error)
}
