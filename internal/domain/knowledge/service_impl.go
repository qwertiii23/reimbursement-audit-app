package knowledge

import (
	"context"
	"errors"
	"fmt"
	"reimbursement-audit/internal/domain/user"

	"github.com/google/uuid"
)

type KnowledgeServiceImpl struct {
	repo           KnowledgeRepository
	userRepository user.Repository
}

func NewKnowledgeService(repo KnowledgeRepository, userRepo user.Repository) KnowledgeService {
	return &KnowledgeServiceImpl{
		repo:           repo,
		userRepository: userRepo,
	}
}

func (s *KnowledgeServiceImpl) UploadFile(file *KnowledgeFile) error {
	if file.ID == "" {
		file.ID = uuid.New().String()
	}
	file.Status = "active"
	return s.repo.Create(file)
}

func (s *KnowledgeServiceImpl) GetFileByID(id string) (*KnowledgeFile, error) {
	return s.repo.GetByID(id)
}

func (s *KnowledgeServiceImpl) GetAllFiles(filter *KnowledgeFileFilter) ([]*KnowledgeFile, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	return s.repo.GetAll(filter)
}

func (s *KnowledgeServiceImpl) UpdateFile(file *KnowledgeFile) error {
	if file.ID == "" {
		return errors.New("文件ID不能为空")
	}
	return s.repo.Update(file)
}

func (s *KnowledgeServiceImpl) DeleteFile(id string, userID string) error {
	file, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	uploader, err := s.userRepository.GetUserByID(context.Background(), userID)
	if err != nil {
		return err
	}

	if uploader.Role != "admin" && file.UploadedBy != userID {
		return errors.New("无权删除该文件")
	}

	return s.repo.Delete(id)
}

func (s *KnowledgeServiceImpl) DownloadFile(id string) (*KnowledgeFile, error) {
	file, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if err := s.repo.IncrementDownloadCount(id); err != nil {
		fmt.Printf("更新下载次数失败: %v", err)
	}

	return file, nil
}

func (s *KnowledgeServiceImpl) GetFilesByUploader(userID string, filter *KnowledgeFileFilter) ([]*KnowledgeFile, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	return s.repo.GetByUploader(userID, filter)
}
