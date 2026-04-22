package mysql

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"reimbursement-audit/internal/domain/audit"
	"reimbursement-audit/internal/pkg/logger"

	"gorm.io/gorm"
)

type AuditResultModel struct {
	ID              string     `gorm:"column:id;primaryKey;type:varchar(36)"`
	ReimbursementID string     `gorm:"column:reimbursement_id;type:varchar(36);index"`
	Status          string     `gorm:"column:status;type:varchar(20)"`
	WorkflowStatus  string     `gorm:"column:workflow_status;type:varchar(32)"`
	RulePass        bool       `gorm:"column:rule_pass;type:tinyint(1)"`
	RAGPass         bool       `gorm:"column:rag_pass;type:tinyint(1)"`
	FinalPass       bool       `gorm:"column:final_pass;type:tinyint(1)"`
	RiskScore       float64    `gorm:"column:risk_score;type:decimal(5,4)"`
	RiskLevel       string     `gorm:"column:risk_level;type:varchar(20)"`
	Reason          string     `gorm:"column:reason;type:text"`
	RuleResults     string     `gorm:"column:rule_results;type:longtext"`
	RAGResults      string     `gorm:"column:rag_results;type:longtext"`
	Suggestions     string     `gorm:"column:suggestions;type:text"`
	StartedAt       time.Time  `gorm:"column:started_at;type:datetime"`
	CompletedAt     *time.Time `gorm:"column:completed_at;type:datetime"`
	Duration        int64      `gorm:"column:duration;type:bigint"`
	CreatedAt       time.Time  `gorm:"column:created_at;type:datetime"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;type:datetime"`
}

func (AuditResultModel) TableName() string {
	return "audit_results"
}

type AuditRepository struct {
	client *Client
	logger logger.Logger
}

func NewAuditRepository(client *Client, logger logger.Logger) audit.Repository {
	return &AuditRepository{
		client: client,
		logger: logger,
	}
}

func (r *AuditRepository) CreateAudit(ctx context.Context, audit *audit.AuditResult) error {
	model := r.toModel(audit)

	db := r.client.GetDB()
	if err := db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.WithContext(ctx).Error("创建审核记录失败",
			logger.NewField("error", err.Error()),
			logger.NewField("audit_id", audit.ID))
		return err
	}

	r.logger.WithContext(ctx).Info("创建审核记录成功",
		logger.NewField("audit_id", audit.ID))
	return nil
}

func (r *AuditRepository) GetAuditByID(ctx context.Context, id string) (*audit.AuditResult, error) {
	db := r.client.GetDB()
	var model AuditResultModel
	if err := db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("审核记录不存在",
				logger.NewField("audit_id", id))
			return nil, errors.New("审核记录不存在")
		}
		r.logger.WithContext(ctx).Error("获取审核记录失败",
			logger.NewField("error", err.Error()),
			logger.NewField("audit_id", id))
		return nil, err
	}

	return r.toDomain(&model), nil
}

func (r *AuditRepository) GetAuditByReimbursementID(ctx context.Context, reimbursementID string) (*audit.AuditResult, error) {
	db := r.client.GetDB()
	var model AuditResultModel
	if err := db.WithContext(ctx).Where("reimbursement_id = ?", reimbursementID).Order("created_at DESC").First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("审核记录不存在",
				logger.NewField("reimbursement_id", reimbursementID))
			return nil, errors.New("审核记录不存在")
		}
		r.logger.WithContext(ctx).Error("获取审核记录失败",
			logger.NewField("error", err.Error()),
			logger.NewField("reimbursement_id", reimbursementID))
		return nil, err
	}

	return r.toDomain(&model), nil
}

func (r *AuditRepository) UpdateAudit(ctx context.Context, audit *audit.AuditResult) error {
	model := r.toModel(audit)

	db := r.client.GetDB()
	if err := db.WithContext(ctx).Save(model).Error; err != nil {
		r.logger.WithContext(ctx).Error("更新审核记录失败",
			logger.NewField("error", err.Error()),
			logger.NewField("audit_id", audit.ID))
		return err
	}

	r.logger.WithContext(ctx).Info("更新审核记录成功",
		logger.NewField("audit_id", audit.ID))
	return nil
}

func (r *AuditRepository) ListAudits(ctx context.Context, filter *audit.AuditFilter) ([]*audit.AuditResult, int64, error) {
	db := r.client.GetDB()
	query := db.Model(&AuditResultModel{})

	if filter != nil {
		if filter.ReimbursementID != "" {
			query = query.Where("reimbursement_id = ?", filter.ReimbursementID)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.StartTime != nil {
			query = query.Where("created_at >= ?", *filter.StartTime)
		}
		if filter.EndTime != nil {
			query = query.Where("created_at <= ?", *filter.EndTime)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		r.logger.WithContext(ctx).Error("统计审核记录数量失败",
			logger.NewField("error", err.Error()))
		return nil, 0, err
	}

	if filter != nil && filter.Page > 0 && filter.Size > 0 {
		offset := (filter.Page - 1) * filter.Size
		query = query.Offset(offset).Limit(filter.Size)
	}

	var models []AuditResultModel
	if err := query.Order("created_at DESC").Find(&models).Error; err != nil {
		r.logger.WithContext(ctx).Error("查询审核记录列表失败",
			logger.NewField("error", err.Error()))
		return nil, 0, err
	}

	results := make([]*audit.AuditResult, len(models))
	for i, model := range models {
		results[i] = r.toDomain(&model)
	}

	return results, total, nil
}

func (r *AuditRepository) DeleteAudit(ctx context.Context, id string) error {
	db := r.client.GetDB()
	if err := db.WithContext(ctx).Delete(&AuditResultModel{}, "id = ?", id).Error; err != nil {
		r.logger.WithContext(ctx).Error("删除审核记录失败",
			logger.NewField("error", err.Error()),
			logger.NewField("audit_id", id))
		return err
	}

	r.logger.WithContext(ctx).Info("删除审核记录成功",
		logger.NewField("audit_id", id))
	return nil
}

func (r *AuditRepository) toModel(domain *audit.AuditResult) *AuditResultModel {
	ruleResultsJSON, _ := json.Marshal(domain.RuleResults)
	ragResultsJSON, _ := json.Marshal(domain.RAGResults)
	suggestionsJSON, _ := json.Marshal(domain.Suggestions)

	return &AuditResultModel{
		ID:              domain.ID,
		ReimbursementID: domain.ReimbursementID,
		Status:          string(domain.Status),
		WorkflowStatus:  string(domain.WorkflowStatus),
		RulePass:        domain.RulePass,
		RAGPass:         domain.RAGPass,
		FinalPass:       domain.FinalPass,
		RiskScore:       domain.RiskScore,
		RiskLevel:       domain.RiskLevel,
		Reason:          domain.Reason,
		RuleResults:     string(ruleResultsJSON),
		RAGResults:      string(ragResultsJSON),
		Suggestions:     string(suggestionsJSON),
		StartedAt:       domain.StartedAt,
		CompletedAt:     domain.CompletedAt,
		Duration:        domain.Duration,
		CreatedAt:       domain.CreatedAt,
		UpdatedAt:       domain.UpdatedAt,
	}
}

func (r *AuditRepository) toDomain(model *AuditResultModel) *audit.AuditResult {
	var ruleResults []*audit.RuleValidationResult
	if model.RuleResults != "" {
		json.Unmarshal([]byte(model.RuleResults), &ruleResults)
	}

	var ragResults *audit.RAGAnalysisResult
	if model.RAGResults != "" {
		json.Unmarshal([]byte(model.RAGResults), &ragResults)
	}

	var suggestions []string
	if model.Suggestions != "" {
		json.Unmarshal([]byte(model.Suggestions), &suggestions)
	}

	return &audit.AuditResult{
		ID:              model.ID,
		ReimbursementID: model.ReimbursementID,
		Status:          audit.AuditStatus(model.Status),
		WorkflowStatus:  audit.WorkflowStatus(model.WorkflowStatus),
		RulePass:        model.RulePass,
		RAGPass:         model.RAGPass,
		FinalPass:       model.FinalPass,
		RiskScore:       model.RiskScore,
		RiskLevel:       model.RiskLevel,
		Reason:          model.Reason,
		RuleResults:     ruleResults,
		RAGResults:      ragResults,
		Suggestions:     suggestions,
		StartedAt:       model.StartedAt,
		CompletedAt:     model.CompletedAt,
		Duration:        model.Duration,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}

type AuditFlowLogModel struct {
	ID              string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	ReimbursementID string    `gorm:"column:reimbursement_id;type:varchar(36);index:idx_reimbursement_id"`
	AuditID         string    `gorm:"column:audit_id;type:varchar(36);index:idx_audit_id"`
	FlowStatus      string    `gorm:"column:flow_status;type:varchar(32);index:idx_flow_status"`
	FlowType        string    `gorm:"column:flow_type;type:varchar(16)"`
	OperatorID      *string   `gorm:"column:operator_id;type:varchar(36)"`
	OperatorName    *string   `gorm:"column:operator_name;type:varchar(64)"`
	Action          string    `gorm:"column:action;type:varchar(32)"`
	Reason          *string   `gorm:"column:reason;type:text"`
	Result          *string   `gorm:"column:result;type:text"`
	IPAddress       *string   `gorm:"column:ip_address;type:varchar(64)"`
	CreatedAt       time.Time `gorm:"column:created_at;type:datetime;index:idx_created_at"`
}

func (AuditFlowLogModel) TableName() string {
	return "audit_flow_logs"
}

func (r *AuditRepository) CreateFlowLog(ctx context.Context, flowLog *audit.AuditFlowLog) error {
	model := r.toFlowLogModel(flowLog)

	db := r.client.GetDB()
	if err := db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.WithContext(ctx).Error("创建流程日志失败",
			logger.NewField("error", err.Error()),
			logger.NewField("flow_log_id", flowLog.ID))
		return err
	}

	r.logger.WithContext(ctx).Info("创建流程日志成功",
		logger.NewField("flow_log_id", flowLog.ID))
	return nil
}

func (r *AuditRepository) GetFlowLogsByReimbursementID(ctx context.Context, reimbursementID string) ([]*audit.AuditFlowLog, error) {
	db := r.client.GetDB()
	var models []AuditFlowLogModel
	if err := db.WithContext(ctx).Where("reimbursement_id = ?", reimbursementID).Order("created_at ASC").Find(&models).Error; err != nil {
		r.logger.WithContext(ctx).Error("获取流程日志失败",
			logger.NewField("error", err.Error()),
			logger.NewField("reimbursement_id", reimbursementID))
		return nil, err
	}

	flowLogs := make([]*audit.AuditFlowLog, len(models))
	for i, model := range models {
		flowLogs[i] = r.toFlowLogDomain(&model)
	}

	return flowLogs, nil
}

func (r *AuditRepository) GetFlowLogsByAuditID(ctx context.Context, auditID string) ([]*audit.AuditFlowLog, error) {
	db := r.client.GetDB()
	var models []AuditFlowLogModel
	if err := db.WithContext(ctx).Where("audit_id = ?", auditID).Order("created_at ASC").Find(&models).Error; err != nil {
		r.logger.WithContext(ctx).Error("获取流程日志失败",
			logger.NewField("error", err.Error()),
			logger.NewField("audit_id", auditID))
		return nil, err
	}

	flowLogs := make([]*audit.AuditFlowLog, len(models))
	for i, model := range models {
		flowLogs[i] = r.toFlowLogDomain(&model)
	}

	return flowLogs, nil
}

func (r *AuditRepository) toFlowLogModel(domain *audit.AuditFlowLog) *AuditFlowLogModel {
	return &AuditFlowLogModel{
		ID:              domain.ID,
		ReimbursementID: domain.ReimbursementID,
		AuditID:         domain.AuditID,
		FlowStatus:      string(domain.FlowStatus),
		FlowType:        string(domain.FlowType),
		OperatorID:      domain.OperatorID,
		OperatorName:    domain.OperatorName,
		Action:          string(domain.Action),
		Reason:          domain.Reason,
		Result:          domain.Result,
		IPAddress:       domain.IPAddress,
		CreatedAt:       domain.CreatedAt,
	}
}

func (r *AuditRepository) toFlowLogDomain(model *AuditFlowLogModel) *audit.AuditFlowLog {
	return &audit.AuditFlowLog{
		ID:              model.ID,
		ReimbursementID: model.ReimbursementID,
		AuditID:         model.AuditID,
		FlowStatus:      audit.FlowStatus(model.FlowStatus),
		FlowType:        audit.FlowType(model.FlowType),
		OperatorID:      model.OperatorID,
		OperatorName:    model.OperatorName,
		Action:          audit.FlowAction(model.Action),
		Reason:          model.Reason,
		Result:          model.Result,
		IPAddress:       model.IPAddress,
		CreatedAt:       model.CreatedAt,
	}
}
