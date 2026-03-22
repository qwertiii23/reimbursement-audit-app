package response

import "time"

// UpdateReimbursementResponse 报销单修改响应
type UpdateReimbursementResponse struct {
	ReimbursementID string    `json:"reimbursement_id"` // 报销单ID
	UserID          string    `json:"user_id"`          // 用户ID
	UserName        string    `json:"user_name"`        // 用户姓名
	Type            string    `json:"type"`             // 报销类型
	Title           string    `json:"title"`            // 报销标题
	Description     string    `json:"description"`      // 报销描述
	TotalAmount     float64   `json:"total_amount"`     // 总金额
	Currency        string    `json:"currency"`         // 币种
	ApplyDate       time.Time `json:"apply_date"`       // 申请日期
	ExpenseDate     time.Time `json:"expense_date"`     // 费用发生日期
	Status          string    `json:"status"`           // 状态
	UpdatedAt       time.Time `json:"updated_at"`       // 更新时间
}

// NewUpdateReimbursementResponse 创建报销单修改响应
func NewUpdateReimbursementResponse(
	reimbursementID, userID, userName, typ, title, description string,
	totalAmount float64, currency string,
	applyDate, expenseDate time.Time,
	status string,
	updatedAt time.Time,
) *UpdateReimbursementResponse {
	return &UpdateReimbursementResponse{
		ReimbursementID: reimbursementID,
		UserID:          userID,
		UserName:        userName,
		Type:            typ,
		Title:           title,
		Description:     description,
		TotalAmount:     totalAmount,
		Currency:        currency,
		ApplyDate:       applyDate,
		ExpenseDate:     expenseDate,
		Status:          status,
		UpdatedAt:       updatedAt,
	}
}
