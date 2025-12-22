// service.go 报销单领域服务
// 功能点：
// 1. 实现报销单相关的业务规则
// 2. 处理发票相关的业务逻辑
// 3. 提供领域模型验证
// 4. 封装复杂的业务计算

package reimbursement

import (
	"context"
	"errors"
	"time"

	"reimbursement-audit/internal/domain/ocr"
	"reimbursement-audit/internal/pkg/logger"

	"github.com/google/uuid"
)

// Service 报销单领域服务接口
type Service interface {
	// CreateReimbursement 创建报销单
	CreateReimbursement(ctx context.Context, req *CreateReimbursementRequest) (*Reimbursement, error)

	// ValidateReimbursement 验证报销单
	ValidateReimbursement(ctx context.Context, reimbursement *Reimbursement) error

	// ValidateInvoice 验证发票
	ValidateInvoice(ctx context.Context, invoice *ocr.Invoice) error
}

// CreateReimbursementRequest 创建报销单请求
type CreateReimbursementRequest struct {
	UserID      string  `json:"user_id"`
	UserName    string  `json:"user_name"`
	Department  string  `json:"department"`
	Category    string  `json:"category"`
	Reason      string  `json:"reason"`
	Description string  `json:"description"`
	TotalAmount float64 `json:"total_amount"`
	ApplyDate   string  `json:"apply_date"`
	ExpenseDate string  `json:"expense_date"`
}

// DomainService 报销单领域服务实现
type DomainService struct {
	repo   Repository
	logger logger.Logger
}

// NewDomainService 创建报销单领域服务
func NewDomainService(repo Repository, logger logger.Logger) Service {
	return &DomainService{
		repo:   repo,
		logger: logger,
	}
}

// CreateReimbursement 创建报销单
func (s *DomainService) CreateReimbursement(ctx context.Context, req *CreateReimbursementRequest) (*Reimbursement, error) {
	// 基本参数验证
	if req.UserID == "" {
		return nil, errors.New("用户ID不能为空")
	}
	if req.TotalAmount <= 0 {
		return nil, errors.New("报销金额必须大于0")
	}

	// 解析日期
	applyDate, expenseDate, err := s.parseDates(ctx, req.ApplyDate, req.ExpenseDate)
	if err != nil {
		s.logger.WithContext(ctx).Error("日期解析失败",
			logger.NewField("error", err.Error()),
			logger.NewField("apply_date", req.ApplyDate),
			logger.NewField("expense_date", req.ExpenseDate))
		return nil, err
	}

	// 创建报销单领域模型
	now := time.Now()
	reimbursement := &Reimbursement{
		ID:          uuid.New().String(),
		UserID:      req.UserID,
		UserName:    req.UserName,
		Department:  req.Department,
		Type:        req.Category, // 使用Category作为Type
		Title:       req.Reason,   // 使用Reason作为Title
		Description: req.Description,
		TotalAmount: req.TotalAmount,
		Currency:    "CNY", // 默认使用人民币
		ApplyDate:   applyDate,
		ExpenseDate: expenseDate,
		Status:      "待提交", // 初始状态为"待提交"
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 验证报销单
	if err := s.ValidateReimbursement(ctx, reimbursement); err != nil {
		s.logger.WithContext(ctx).Error("报销单验证失败",
			logger.NewField("error", err.Error()),
			logger.NewField("reimbursement_id", reimbursement.ID))
		return nil, err
	}

	// 保存到数据库
	if err := s.repo.CreateReimbursement(ctx, reimbursement); err != nil {
		s.logger.WithContext(ctx).Error("保存报销单失败",
			logger.NewField("error", err.Error()),
			logger.NewField("reimbursement_id", reimbursement.ID))
		return nil, err
	}

	return reimbursement, nil
}

// ValidateReimbursement 验证报销单
func (s *DomainService) ValidateReimbursement(ctx context.Context, reimbursement *Reimbursement) error {
	// 基本字段验证
	if reimbursement.ID == "" {
		return errors.New("报销单ID不能为空")
	}
	if reimbursement.UserID == "" {
		return errors.New("用户ID不能为空")
	}
	if reimbursement.TotalAmount <= 0 {
		return errors.New("报销金额必须大于0")
	}

	// 业务规则验证
	if reimbursement.ApplyDate.After(time.Now()) {
		return errors.New("申请日期不能是未来日期")
	}

	if reimbursement.ExpenseDate.After(time.Now()) {
		return errors.New("费用发生日期不能是未来日期")
	}

	if reimbursement.ExpenseDate.After(reimbursement.ApplyDate) {
		return errors.New("费用发生日期不能晚于申请日期")
	}

	// 可以添加更多业务规则验证...

	return nil
}

// ValidateInvoice 验证发票
func (s *DomainService) ValidateInvoice(ctx context.Context, invoice *ocr.Invoice) error {
	// 基本字段验证
	if invoice.ID == "" {
		return errors.New("发票ID不能为空")
	}
	if invoice.ReimbursementID == "" {
		return errors.New("报销单ID不能为空")
	}
	if invoice.ImagePath == "" {
		return errors.New("发票图片路径不能为空")
	}

	// 可以添加更多业务规则验证...

	return nil
}

// parseDates 解析申请日期和费用发生日期
func (s *DomainService) parseDates(ctx context.Context, applyDateStr, expenseDateStr string) (time.Time, time.Time, error) {
	var applyDate, expenseDate time.Time
	var err error

	// 如果提供了申请日期，解析它
	if applyDateStr != "" {
		applyDate, err = time.Parse("2006-01-02", applyDateStr)
		if err != nil {
			s.logger.WithContext(ctx).Error("申请日期格式不正确",
				logger.NewField("error", err.Error()),
				logger.NewField("apply_date", applyDateStr))
			return time.Time{}, time.Time{}, errors.New("申请日期格式不正确，应为YYYY-MM-DD")
		}
	} else {
		// 如果没有提供申请日期，使用当前日期
		applyDate = time.Now()
	}

	// 如果提供了费用发生日期，解析它
	if expenseDateStr != "" {
		expenseDate, err = time.Parse("2006-01-02", expenseDateStr)
		if err != nil {
			s.logger.WithContext(ctx).Error("费用发生日期格式不正确",
				logger.NewField("error", err.Error()),
				logger.NewField("expense_date", expenseDateStr))
			return time.Time{}, time.Time{}, errors.New("费用发生日期格式不正确，应为YYYY-MM-DD")
		}
	} else {
		// 如果没有提供费用发生日期，使用申请日期
		expenseDate = applyDate
	}

	return applyDate, expenseDate, nil
}
