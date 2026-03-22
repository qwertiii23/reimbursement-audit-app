package mysql

import (
	"context"
	"errors"
	"time"

	"reimbursement-audit/internal/domain/user"
	"reimbursement-audit/internal/pkg/logger"

	"gorm.io/gorm"
)

type UserRepository struct {
	client *Client
	logger logger.Logger
}

func NewUserRepository(client *Client, logger logger.Logger) user.Repository {
	return &UserRepository{
		client: client,
		logger: logger,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *user.User) error {
	exists, err := r.CheckUsernameExists(ctx, user.Username)
	if err != nil {
		r.logger.WithContext(ctx).Error("检查用户名唯一性失败", logger.NewField("error", err.Error()), logger.NewField("username", user.Username))
		return err
	}
	if exists {
		r.logger.WithContext(ctx).Warn("用户名已存在", logger.NewField("username", user.Username))
		return errors.New("用户名已存在")
	}

	if user.Email != "" {
		exists, err := r.CheckEmailExists(ctx, user.Email)
		if err != nil {
			r.logger.WithContext(ctx).Error("检查邮箱唯一性失败", logger.NewField("error", err.Error()), logger.NewField("email", user.Email))
			return err
		}
		if exists {
			r.logger.WithContext(ctx).Warn("邮箱已存在", logger.NewField("email", user.Email))
			return errors.New("邮箱已存在")
		}
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	result := r.client.GetDB().WithContext(ctx).Create(user)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("创建用户失败", logger.NewField("error", result.Error.Error()), logger.NewField("username", user.Username))
		return result.Error
	}

	r.logger.WithContext(ctx).Info("创建用户成功", logger.NewField("user_id", user.ID), logger.NewField("username", user.Username))
	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	var u user.User
	result := r.client.GetDB().WithContext(ctx).Where("id = ?", id).First(&u)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("用户不存在", logger.NewField("user_id", id))
			return nil, result.Error
		}
		r.logger.WithContext(ctx).Error("获取用户失败", logger.NewField("error", result.Error.Error()), logger.NewField("user_id", id))
		return nil, result.Error
	}
	return &u, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*user.User, error) {
	var u user.User
	result := r.client.GetDB().WithContext(ctx).Where("username = ?", username).First(&u)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("用户不存在", logger.NewField("username", username))
			return nil, result.Error
		}
		r.logger.WithContext(ctx).Error("获取用户失败", logger.NewField("error", result.Error.Error()), logger.NewField("username", username))
		return nil, result.Error
	}
	return &u, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	result := r.client.GetDB().WithContext(ctx).Where("email = ?", email).First(&u)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("用户不存在", logger.NewField("email", email))
			return nil, result.Error
		}
		r.logger.WithContext(ctx).Error("获取用户失败", logger.NewField("error", result.Error.Error()), logger.NewField("email", email))
		return nil, result.Error
	}
	return &u, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *user.User) error {
	user.UpdatedAt = time.Now()
	result := r.client.GetDB().WithContext(ctx).Model(user).Where("id = ?", user.ID).Updates(user)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("更新用户失败", logger.NewField("error", result.Error.Error()), logger.NewField("user_id", user.ID))
		return result.Error
	}
	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("用户不存在，更新失败", logger.NewField("user_id", user.ID))
		return errors.New("用户不存在")
	}
	r.logger.WithContext(ctx).Info("更新用户成功", logger.NewField("user_id", user.ID))
	return nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	now := time.Now()
	result := r.client.GetDB().WithContext(ctx).Model(&user.User{}).
		Where("id = ?", userID).
		Update("last_login", now)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("更新最后登录时间失败", logger.NewField("error", result.Error.Error()), logger.NewField("user_id", userID))
		return result.Error
	}
	return nil
}

func (r *UserRepository) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.client.GetDB().WithContext(ctx).Model(&user.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.client.GetDB().WithContext(ctx).Model(&user.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
