// model.go 报销单领域模型
// 功能点：
// 1. 定义报销单数据模型
// 2. 定义发票数据模型
// 3. 定义审核结果数据模型
// 4. 定义审核状态数据模型
// 5. 定义校验结果数据模型
// 6. 提供模型转换和验证方法

package reimbursement

import (
	"time"

	"reimbursement-audit/internal/domain/ocr"
)

// Reimbursement 报销单模型
type Reimbursement struct {
	ID               string         `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`                                    // 报销单ID
	UserID           string         `json:"user_id" gorm:"type:varchar(36);not null;column:user_id"`                            // 用户ID
	UserName         string         `json:"user_name" gorm:"type:varchar(100);not null;column:user_name"`                       // 用户姓名
	Department       string         `json:"department" gorm:"type:varchar(100);column:department"`                              // 所属部门
	ApplicantLevel   string         `json:"applicant_level" gorm:"type:varchar(20);column:applicant_level"`                     // 申请人级别(高管/经理/员工)
	Type             string         `json:"type" gorm:"type:varchar(50);column:type"`                                           // 报销类型(交通/住宿/餐饮等)
	Title            string         `json:"title" gorm:"type:varchar(200);not null;column:title"`                               // 报销标题
	Description      string         `json:"description" gorm:"type:text;column:description"`                                    // 报销描述
	TotalAmount      float64        `json:"total_amount" gorm:"type:decimal(10,2);not null;column:total_amount"`                // 总金额
	Currency         string         `json:"currency" gorm:"type:varchar(10);default:'CNY';column:currency"`                     // 币种
	ApplyDate        time.Time      `json:"apply_date" gorm:"type:date;not null;column:apply_date"`                             // 申请日期
	ExpenseDate      time.Time      `json:"expense_date" gorm:"type:date;column:expense_date"`                                  // 费用发生日期
	StartDate        *time.Time     `json:"start_date" gorm:"type:date;column:start_date"`                                      // 出差开始日期
	EndDate          *time.Time     `json:"end_date" gorm:"type:date;column:end_date"`                                          // 出差结束日期
	Destination      string         `json:"destination" gorm:"type:varchar(100);column:destination"`                            // 出差目的地
	City             string         `json:"city" gorm:"type:varchar(50);column:city"`                                           // 出差城市
	Province         string         `json:"province" gorm:"type:varchar(50);column:province"`                                   // 出差省份
	TravelReason     string         `json:"travel_reason" gorm:"type:varchar(200);column:travel_reason"`                        // 出差事由
	Transportation   string         `json:"transportation" gorm:"type:varchar(50);column:transportation"`                       // 交通工具
	ProjectCode      string         `json:"project_code" gorm:"type:varchar(50);column:project_code"`                           // 项目编码
	BudgetCode       string         `json:"budget_code" gorm:"type:varchar(50);column:budget_code"`                             // 预算科目
	ApprovalRequired bool           `json:"approval_required" gorm:"type:boolean;default:false;column:approval_required"`       // 是否需要审批
	ApprovedBy       string         `json:"approved_by" gorm:"type:varchar(36);column:approved_by"`                             // 审批人ID
	ApprovedAt       *time.Time     `json:"approved_at" gorm:"type:datetime;column:approved_at"`                                // 审批时间
	Invoices         []*ocr.Invoice `json:"invoices" gorm:"foreignKey:ReimbursementID;constraint:OnDelete:CASCADE"`             // 发票列表
	Status           string         `json:"status" gorm:"type:varchar(20);not null;default:'pending_submission';column:status"` // 状态(待提交/待审核/审核中/已完成/已驳回)
	CreatedAt        time.Time      `json:"created_at" gorm:"autoCreateTime"`                                                   // 创建时间
	UpdatedAt        time.Time      `json:"updated_at" gorm:"autoUpdateTime"`                                                   // 更新时间
	// AuditResults []*AuditResult `json:"audit_results" gorm:"foreignKey:ReimbursementID;constraint:OnDelete:CASCADE"` // 审核结果列表
}

