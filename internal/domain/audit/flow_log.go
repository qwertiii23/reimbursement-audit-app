package audit

import (
	"time"
)

// FlowType 流程类型
type FlowType string

const (
	FlowTypeIntelligent FlowType = "intelligent"
	FlowTypeManual      FlowType = "manual"
)

// FlowStatus 流程状态
type FlowStatus string

const (
	FlowStatusIntelligentStart FlowStatus = "智能审核开始"
	FlowStatusIntelligentPass  FlowStatus = "智能审核通过"
	FlowStatusIntelligentFail  FlowStatus = "智能审核失败"
	FlowStatusManualStart      FlowStatus = "人工审核开始"
	FlowStatusManualPass       FlowStatus = "人工审核通过"
	FlowStatusManualReject     FlowStatus = "人工审核驳回"
	FlowStatusWithdrawn        FlowStatus = "已撤回"
)

// FlowAction 流程动作
type FlowAction string

const (
	FlowActionStartAudit    FlowAction = "start_audit"
	FlowActionPassAudit     FlowAction = "pass_audit"
	FlowActionRejectAudit   FlowAction = "reject_audit"
	FlowActionWithdrawAudit FlowAction = "withdraw_audit"
)

// AuditFlowLog 审核流程日志
type AuditFlowLog struct {
	ID              string     `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`
	ReimbursementID string     `json:"reimbursement_id" gorm:"type:varchar(36);not null;index:idx_reimbursement_id;column:reimbursement_id"`
	AuditID         string     `json:"audit_id" gorm:"type:varchar(36);not null;index:idx_audit_id;column:audit_id"`
	FlowStatus      FlowStatus `json:"flow_status" gorm:"type:varchar(32);not null;index:idx_flow_status;column:flow_status"`
	FlowType        FlowType   `json:"flow_type" gorm:"type:varchar(16);not null;column:flow_type"`
	OperatorID      *string    `json:"operator_id" gorm:"type:varchar(36);column:operator_id"`
	OperatorName    *string    `json:"operator_name" gorm:"type:varchar(64);column:operator_name"`
	Action          FlowAction `json:"action" gorm:"type:varchar(32);not null;column:action"`
	Reason          *string    `json:"reason" gorm:"type:text;column:reason"`
	Result          *string    `json:"result" gorm:"type:text;column:result"`
	IPAddress       *string    `json:"ip_address" gorm:"type:varchar(64);column:ip_address"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime;index:idx_created_at;column:created_at"`
}

// TableName 指定表名
func (AuditFlowLog) TableName() string {
	return "audit_flow_logs"
}
