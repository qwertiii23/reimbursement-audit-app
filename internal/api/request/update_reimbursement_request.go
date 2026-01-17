package request

import (
	"errors"
	"strings"
	"time"
)

// UpdateReimbursementRequest 报销单修改请求
type UpdateReimbursementRequest struct {
	ID               string  `json:"id" binding:"required"`                 // 报销单ID，必填
	Type             *string `json:"type"`                                  // 报销类型
	Title            *string `json:"title"`                                 // 报销标题
	Description      *string `json:"description"`                           // 报销描述
	TotalAmount      *float64 `json:"total_amount"`                         // 总金额
	Currency         *string `json:"currency"`                              // 币种
	ApplyDate        *string `json:"apply_date"`                             // 申请日期，格式：YYYY-MM-DD
	ExpenseDate      *string `json:"expense_date"`                          // 费用发生日期，格式：YYYY-MM-DD
	StartDate        *string `json:"start_date"`                            // 出差开始日期，格式：YYYY-MM-DD
	EndDate          *string `json:"end_date"`                              // 出差结束日期，格式：YYYY-MM-DD
	Destination      *string `json:"destination"`                           // 出差目的地
	City             *string `json:"city"`                                  // 出差城市
	Province         *string `json:"province"`                              // 出差省份
	TravelReason     *string `json:"travel_reason"`                         // 出差事由
	Transportation   *string `json:"transportation"`                         // 交通工具
	ProjectCode      *string `json:"project_code"`                          // 项目编码
	BudgetCode       *string `json:"budget_code"`                           // 预算科目
}

// Validate 校验报销单修改请求
func (r *UpdateReimbursementRequest) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return errors.New("报销单ID不能为空")
	}

	if r.TotalAmount != nil && *r.TotalAmount <= 0 {
		return errors.New("总金额必须大于0")
	}

	if r.ApplyDate != nil {
		if _, err := time.Parse("2006-01-02", *r.ApplyDate); err != nil {
			return errors.New("申请日期格式不正确，应为YYYY-MM-DD")
		}
	}

	if r.ExpenseDate != nil {
		if _, err := time.Parse("2006-01-02", *r.ExpenseDate); err != nil {
			return errors.New("费用发生日期格式不正确，应为YYYY-MM-DD")
		}
	}

	if r.StartDate != nil {
		if _, err := time.Parse("2006-01-02", *r.StartDate); err != nil {
			return errors.New("出差开始日期格式不正确，应为YYYY-MM-DD")
		}
	}

	if r.EndDate != nil {
		if _, err := time.Parse("2006-01-02", *r.EndDate); err != nil {
			return errors.New("出差结束日期格式不正确，应为YYYY-MM-DD")
		}
	}

	return nil
}
