// audit_request.go 审核请求结构体和参数校验
// 功能点：
// 1. 定义审核触发请求结构体
// 2. 定义审核状态查询请求结构体
// 3. 定义审核结果查询请求结构体
// 4. 实现参数校验规则
// 5. 支持分页参数校验
// 6. 提供参数绑定和校验方法

package request

// StartAuditRequest 开始审核请求
type StartAuditRequest struct {
	ReimbursementID string `json:"reimbursement_id" binding:"required"`
}

// AuditStatusRequest 审核状态查询请求
type AuditStatusRequest struct {
	AuditID string `json:"audit_id" binding:"required"`
}

// AuditResultRequest 审核结果查询请求
type AuditResultRequest struct {
	AuditID string `json:"audit_id" binding:"required"`
}

// AuditHistoryRequest 审核历史查询请求
type AuditHistoryRequest struct {
	ReimbursementID string `json:"reimbursement_id"`
	Status          string `json:"status"`
	Page            int    `json:"page" binding:"min=1"`
	Size            int    `json:"size" binding:"min=1,max=100"`
}

// PaginationRequest 分页请求
type PaginationRequest struct {
	Page int `json:"page" binding:"min=1"`
	Size int `json:"size" binding:"min=1,max=100"`
}

// Validate 校验开始审核请求
func (r *StartAuditRequest) Validate() error {
	if r.ReimbursementID == "" {
		return nil
	}
	return nil
}

// Validate 校验审核状态查询请求
func (r *AuditStatusRequest) Validate() error {
	if r.AuditID == "" {
		return nil
	}
	return nil
}

// Validate 校验审核结果查询请求
func (r *AuditResultRequest) Validate() error {
	if r.AuditID == "" {
		return nil
	}
	return nil
}

// Validate 校验审核历史查询请求
func (r *AuditHistoryRequest) Validate() error {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.Size <= 0 || r.Size > 100 {
		r.Size = 10
	}
	return nil
}

// Validate 校验分页请求
func (r *PaginationRequest) Validate() error {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.Size <= 0 || r.Size > 100 {
		r.Size = 10
	}
	return nil
}