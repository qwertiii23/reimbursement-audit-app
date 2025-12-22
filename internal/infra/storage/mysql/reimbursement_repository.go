// mysql_repository.go MySQL仓储实现
// 功能点：
// 1. 实现报销单仓储接口
// 2. 实现发票仓储接口
// 3. 实现审核结果仓储接口
// 4. 提供MySQL数据访问实现
// 5. 支持事务管理
// 6. 支持查询和分页

package mysql

import (
	"context"
	"errors"
	"time"

	"reimbursement-audit/internal/domain/reimbursement"
	"reimbursement-audit/internal/pkg/logger"

	"gorm.io/gorm"
)

// ReimbursementRepository 报销单MySQL仓储实现
type ReimbursementRepository struct {
	client *Client
	logger logger.Logger
}

// NewReimbursementRepository 创建报销单MySQL仓储实例
func NewReimbursementRepository(client *Client, logger logger.Logger) reimbursement.Repository {
	return &ReimbursementRepository{client: client, logger: logger}
}

// CreateReimbursement 创建报销单
func (r *ReimbursementRepository) CreateReimbursement(ctx context.Context, reimbursement *reimbursement.Reimbursement) error {
	// 使用GORM创建报销单记录
	result := r.client.GetDB().WithContext(ctx).Create(reimbursement)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("创建报销单失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("user_id", reimbursement.UserID))
		return result.Error
	}

	return nil
}

// GetReimbursementByID 根据ID获取报销单
func (r *ReimbursementRepository) GetReimbursementByID(ctx context.Context, id string) (*reimbursement.Reimbursement, error) {
	var reimbursement reimbursement.Reimbursement

	// 使用GORM查询报销单
	result := r.client.GetDB().WithContext(ctx).Where("id = ?", id).First(&reimbursement)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("报销单不存在",
				logger.NewField("reimbursement_id", id))
			return nil, result.Error
		}
		r.logger.WithContext(ctx).Error("获取报销单失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("reimbursement_id", id))
		return nil, result.Error
	}

	// 不在此处加载发票列表，保持聚合根的独立性
	// 发票列表应由应用服务在需要时通过OCRRepository单独加载

	return &reimbursement, nil
}

// UpdateReimbursement 更新报销单
func (r *ReimbursementRepository) UpdateReimbursement(ctx context.Context, reimbursement *reimbursement.Reimbursement) error {
	// 使用GORM更新报销单
	result := r.client.GetDB().WithContext(ctx).Model(reimbursement).
		Where("id = ?", reimbursement.ID).
		Updates(map[string]interface{}{
			"user_id":      reimbursement.UserID,
			"user_name":    reimbursement.UserName,
			"department":   reimbursement.Department,
			"type":         reimbursement.Type,
			"title":        reimbursement.Title,
			"description":  reimbursement.Description,
			"total_amount": reimbursement.TotalAmount,
			"currency":     reimbursement.Currency,
			"apply_date":   reimbursement.ApplyDate,
			"expense_date": reimbursement.ExpenseDate,
			"status":       reimbursement.Status,
			"updated_at":   time.Now(),
		})

	if result.Error != nil {
		r.logger.WithContext(ctx).Error("更新报销单失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("reimbursement_id", reimbursement.ID))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("报销单不存在，更新失败",
			logger.NewField("reimbursement_id", reimbursement.ID))
		return result.Error
	}

	return nil
}

// DeleteReimbursement 删除报销单
func (r *ReimbursementRepository) DeleteReimbursement(ctx context.Context, id string) error {
	// 使用GORM删除报销单
	result := r.client.GetDB().WithContext(ctx).Where("id = ?", id).Delete(&reimbursement.Reimbursement{})

	if result.Error != nil {
		r.logger.WithContext(ctx).Error("删除报销单失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("reimbursement_id", id))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("报销单不存在，删除失败",
			logger.NewField("reimbursement_id", id))
		return result.Error
	}

	return nil
}

// GetReimbursementsByUserID 根据用户ID获取报销单列表
func (r *ReimbursementRepository) GetReimbursementsByUserID(ctx context.Context, userID string, limit, offset int) ([]*reimbursement.Reimbursement, error) {
	var reimbursements []*reimbursement.Reimbursement

	// 使用GORM查询报销单列表
	result := r.client.GetDB().WithContext(ctx).
		Where("user_id = ?", userID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&reimbursements)

	if result.Error != nil {
		r.logger.WithContext(ctx).Error("查询用户报销单失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("user_id", userID),
			logger.NewField("limit", limit),
			logger.NewField("offset", offset))
		return nil, result.Error
	}

	return reimbursements, nil
}

// ListReimbursementsByUserID 根据用户ID获取报销单列表
func (r *ReimbursementRepository) ListReimbursementsByUserID(ctx context.Context, userID string, page, size int) ([]*reimbursement.Reimbursement, int64, error) {
	// 获取总数
	var total int64
	countResult := r.client.GetDB().WithContext(ctx).Model(&reimbursement.Reimbursement{}).Where("user_id = ?", userID).Count(&total)
	if countResult.Error != nil {
		r.logger.WithContext(ctx).Error("获取报销单总数失败",
			logger.NewField("error", countResult.Error.Error()),
			logger.NewField("user_id", userID))
		return nil, 0, countResult.Error
	}

	// 获取分页数据
	offset := (page - 1) * size
	var reimbursements []*reimbursement.Reimbursement
	result := r.client.GetDB().WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&reimbursements)

	if result.Error != nil {
		r.logger.WithContext(ctx).Error("获取报销单列表失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("user_id", userID),
			logger.NewField("page", page),
			logger.NewField("size", size))
		return nil, 0, result.Error
	}

	// 不在此处加载发票列表，保持聚合根的独立性
	// 发票列表应由应用服务在需要时通过OCRRepository单独加载

	return reimbursements, total, nil
}

// ListReimbursementsByDateRange 根据日期范围获取报销单列表
func (r *ReimbursementRepository) ListReimbursementsByDateRange(ctx context.Context, startDate, endDate string, page, size int) ([]*reimbursement.Reimbursement, int64, error) {
	// 获取总数
	var total int64
	countResult := r.client.GetDB().WithContext(ctx).Model(&reimbursement.Reimbursement{}).
		Where("apply_date BETWEEN ? AND ?", startDate, endDate).
		Count(&total)

	if countResult.Error != nil {
		r.logger.WithContext(ctx).Error("获取报销单总数失败",
			logger.NewField("error", countResult.Error.Error()),
			logger.NewField("start_date", startDate),
			logger.NewField("end_date", endDate))
		return nil, 0, countResult.Error
	}

	// 获取分页数据
	offset := (page - 1) * size
	var reimbursements []*reimbursement.Reimbursement
	result := r.client.GetDB().WithContext(ctx).
		Where("apply_date BETWEEN ? AND ?", startDate, endDate).
		Order("apply_date DESC").
		Limit(size).
		Offset(offset).
		Find(&reimbursements)

	if result.Error != nil {
		r.logger.WithContext(ctx).Error("获取报销单列表失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("start_date", startDate),
			logger.NewField("end_date", endDate),
			logger.NewField("page", page),
			logger.NewField("size", size))
		return nil, 0, result.Error
	}

	// 不在此处加载发票列表，保持聚合根的独立性
	// 发票列表应由应用服务在需要时通过OCRRepository单独加载

	return reimbursements, total, nil
}

// ListReimbursementsByStatus 根据状态获取报销单列表
func (r *ReimbursementRepository) ListReimbursementsByStatus(ctx context.Context, status string, page, size int) ([]*reimbursement.Reimbursement, int64, error) {
	// 获取总数
	var total int64
	countResult := r.client.GetDB().WithContext(ctx).Model(&reimbursement.Reimbursement{}).
		Where("status = ?", status).
		Count(&total)

	if countResult.Error != nil {
		r.logger.WithContext(ctx).Error("获取报销单总数失败",
			logger.NewField("error", countResult.Error.Error()),
			logger.NewField("status", status))
		return nil, 0, countResult.Error
	}

	// 获取分页数据
	offset := (page - 1) * size
	var reimbursements []*reimbursement.Reimbursement
	result := r.client.GetDB().WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&reimbursements)

	if result.Error != nil {
		r.logger.WithContext(ctx).Error("获取报销单列表失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("status", status),
			logger.NewField("page", page),
			logger.NewField("size", size))
		return nil, 0, result.Error
	}

	// 不在此处加载发票列表，保持聚合根的独立性
	// 发票列表应由应用服务在需要时通过OCRRepository单独加载

	return reimbursements, total, nil
}

// SearchReimbursements 搜索报销单
func (r *ReimbursementRepository) SearchReimbursements(ctx context.Context, keyword string, page, size int) ([]*reimbursement.Reimbursement, int64, error) {
	// 获取总数
	var total int64
	searchPattern := "%" + keyword + "%"
	countResult := r.client.GetDB().WithContext(ctx).Model(&reimbursement.Reimbursement{}).
		Where("user_name LIKE ? OR title LIKE ? OR description LIKE ?", searchPattern, searchPattern, searchPattern).
		Count(&total)

	if countResult.Error != nil {
		r.logger.WithContext(ctx).Error("获取报销单总数失败",
			logger.NewField("error", countResult.Error.Error()),
			logger.NewField("keyword", keyword))
		return nil, 0, countResult.Error
	}

	// 获取分页数据
	offset := (page - 1) * size
	var reimbursements []*reimbursement.Reimbursement
	result := r.client.GetDB().WithContext(ctx).
		Where("user_name LIKE ? OR title LIKE ? OR description LIKE ?", searchPattern, searchPattern, searchPattern).
		Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&reimbursements)

	if result.Error != nil {
		r.logger.WithContext(ctx).Error("获取报销单列表失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("keyword", keyword),
			logger.NewField("page", page),
			logger.NewField("size", size))
		return nil, 0, result.Error
	}

	// 不在此处加载发票列表，保持聚合根的独立性
	// 发票列表应由应用服务在需要时通过OCRRepository单独加载

	return reimbursements, total, nil
}
