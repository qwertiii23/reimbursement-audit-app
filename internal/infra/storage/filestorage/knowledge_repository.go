package filestorage

import (
	"mime"
	"os"
	"path/filepath"
	"reimbursement-audit/internal/domain/knowledge"
	"reimbursement-audit/internal/pkg/logger"
	"strings"
)

type KnowledgeRepository struct {
	basePath string
	logger   logger.Logger
}

func NewKnowledgeRepository(basePath string, logger logger.Logger) *KnowledgeRepository {
	return &KnowledgeRepository{
		basePath: basePath,
		logger:   logger,
	}
}

func (r *KnowledgeRepository) Create(file *knowledge.KnowledgeFile) error {
	filePath := filepath.Join(r.basePath, file.FileName)

	existingFile, err := r.GetByFileName(file.FileName)
	if err == nil && existingFile != nil {
		return r.Update(file)
	}

	if err := os.WriteFile(filePath, []byte(file.Description), 0644); err != nil {
		r.logger.Error("写入文件失败", logger.NewField("error", err), logger.NewField("file_name", file.FileName))
		return err
	}

	r.logger.Info("文件创建成功", logger.NewField("file_name", file.FileName), logger.NewField("path", filePath))
	return nil
}

func (r *KnowledgeRepository) GetByID(id string) (*knowledge.KnowledgeFile, error) {
	return r.GetByFileName(id)
}

func (r *KnowledgeRepository) GetByFileName(fileName string) (*knowledge.KnowledgeFile, error) {
	filePath := filepath.Join(r.basePath, fileName)

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		r.logger.Error("获取文件信息失败", logger.NewField("error", err), logger.NewField("file_name", fileName))
		return nil, err
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		r.logger.Error("读取文件失败", logger.NewField("error", err), logger.NewField("file_name", fileName))
		return nil, err
	}

	mimeType := mime.TypeByExtension(filepath.Ext(fileName))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return &knowledge.KnowledgeFile{
		ID:            fileName,
		FileName:      fileName,
		FilePath:      filePath,
		FileType:      mimeType,
		FileSize:      fileInfo.Size(),
		Category:      r.extractCategory(fileName),
		Description:   string(content),
		UploadedBy:    "system",
		UploaderName:  "系统",
		DownloadCount: 0,
		Status:        "active",
		CreatedAt:     fileInfo.ModTime(),
		UpdatedAt:     fileInfo.ModTime(),
	}, nil
}

func (r *KnowledgeRepository) GetAll(filter *knowledge.KnowledgeFileFilter) ([]*knowledge.KnowledgeFile, int64, error) {
	entries, err := os.ReadDir(r.basePath)
	if err != nil {
		r.logger.Error("读取目录失败", logger.NewField("error", err), logger.NewField("path", r.basePath))
		return nil, 0, err
	}

	var files []*knowledge.KnowledgeFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		file, err := r.GetByFileName(fileName)
		if err != nil {
			r.logger.Warn("获取文件信息失败", logger.NewField("error", err), logger.NewField("file_name", fileName))
			continue
		}

		if filter.Category != "" && file.Category != filter.Category {
			continue
		}

		files = append(files, file)
	}

	total := int64(len(files))

	start := (filter.Page - 1) * filter.PageSize
	end := start + filter.PageSize

	if int64(start) > total {
		return []*knowledge.KnowledgeFile{}, total, nil
	}

	if int64(end) > total {
		end = int(total)
	}

	return files[start:end], total, nil
}

func (r *KnowledgeRepository) Update(file *knowledge.KnowledgeFile) error {
	filePath := filepath.Join(r.basePath, file.FileName)

	if err := os.WriteFile(filePath, []byte(file.Description), 0644); err != nil {
		r.logger.Error("更新文件失败", logger.NewField("error", err), logger.NewField("file_name", file.FileName))
		return err
	}

	r.logger.Info("文件更新成功", logger.NewField("file_name", file.FileName))
	return nil
}

func (r *KnowledgeRepository) Delete(id string) error {
	filePath := filepath.Join(r.basePath, id)

	if err := os.Remove(filePath); err != nil {
		r.logger.Error("删除文件失败", logger.NewField("error", err), logger.NewField("file_name", id))
		return err
	}

	r.logger.Info("文件删除成功", logger.NewField("file_name", id))
	return nil
}

func (r *KnowledgeRepository) IncrementDownloadCount(id string) error {
	return nil
}

func (r *KnowledgeRepository) GetFileContent(id string) (string, error) {
	filePath := filepath.Join(r.basePath, id)

	content, err := os.ReadFile(filePath)
	if err != nil {
		r.logger.Error("读取文件内容失败", logger.NewField("error", err), logger.NewField("file_name", id))
		return "", err
	}

	return string(content), nil
}

func (r *KnowledgeRepository) GetByUploader(userID string, filter *knowledge.KnowledgeFileFilter) ([]*knowledge.KnowledgeFile, int64, error) {
	return r.GetAll(filter)
}

func (r *KnowledgeRepository) extractCategory(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))

	categoryMap := map[string]string{
		".pdf":  "policy",
		".doc":  "policy",
		".docx": "policy",
		".txt":  "policy",
		".xls":  "finance",
		".xlsx": "finance",
		".ppt":  "training",
		".pptx": "training",
		".jpg":  "other",
		".jpeg": "other",
		".png":  "other",
	}

	if category, ok := categoryMap[ext]; ok {
		return category
	}

	return "other"
}
