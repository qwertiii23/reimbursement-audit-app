package user

import (
	"context"
	"errors"

	"reimbursement-audit/internal/pkg/crypto"
	"reimbursement-audit/internal/pkg/logger"
)

type ServiceInterface interface {
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
}

type Service struct {
	repo   Repository
	logger logger.Logger
}

func NewUserService(repo Repository, logger logger.Logger) ServiceInterface {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		s.logger.WithContext(ctx).Error("获取用户失败", logger.NewField("error", err.Error()), logger.NewField("username", req.Username))
		return nil, errors.New("用户名或密码错误")
	}

	if !user.IsActive() {
		s.logger.WithContext(ctx).Warn("用户已被禁用", logger.NewField("username", req.Username))
		return nil, errors.New("用户已被禁用")
	}

	if err := crypto.ComparePassword(user.Password, req.Password); err != nil {
		s.logger.WithContext(ctx).Error("密码验证失败", logger.NewField("error", err.Error()), logger.NewField("username", req.Username))
		return nil, errors.New("用户名或密码错误")
	}

	token, err := crypto.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		s.logger.WithContext(ctx).Error("生成token失败", logger.NewField("error", err.Error()))
		return nil, errors.New("生成token失败")
	}

	if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
		s.logger.WithContext(ctx).Error("更新最后登录时间失败", logger.NewField("error", err.Error()))
	}

	s.logger.WithContext(ctx).Info("用户登录成功", logger.NewField("username", req.Username))

	return &LoginResponse{
		Token: token,
		User:  user.ToUserInfo(),
	}, nil
}

func (s *Service) GetUserByID(ctx context.Context, id string) (*User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *Service) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	return s.repo.GetUserByUsername(ctx, username)
}
