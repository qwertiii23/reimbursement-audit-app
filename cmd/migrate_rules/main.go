package main

import (
	"context"
	"fmt"
	"log"
	"reimbursement-audit/internal/domain/ruleengine"
	"reimbursement-audit/internal/infra/storage/mysql"
	"reimbursement-audit/internal/pkg/logger"

	"github.com/google/uuid"
)

func main() {
	mysqlLogger, err := logger.NewLogger(logger.DefaultConfig())
	if err != nil {
		log.Fatalf("创建日志器失败: %v", err)
	}

	mysqlClient := &mysql.Client{}
	mysqlClient.Connect(context.Background(), &mysql.Config{
		Host:      "localhost",
		Port:      3306,
		Username:  "root",
		Password:  "Ljq20040811.",
		DBName:    "reimbursement_audit",
		Charset:   "utf8mb4",
		ParseTime: true,
		Loc:       "Local",
	})

	repo := mysql.NewRuleEngineRepository(mysqlClient, mysqlLogger)

	ctx := context.Background()

	err = migrateFeatures(ctx, repo)
	if err != nil {
		log.Fatalf("迁移特征失败: %v", err)
	}

	err = migrateRules(ctx, repo)
	if err != nil {
		log.Fatalf("迁移规则失败: %v", err)
	}

	fmt.Println("规则迁移完成！")
}

func migrateFeatures(ctx context.Context, repo *mysql.RuleEngineRepository) error {
	features := []*ruleengine.Feature{
		{
			ID:        uuid.New().String(),
			Name:      "发票分类",
			Code:      "invoice_category",
			Type:      "string",
			ValueType: "single",
			Category:  "invoice",
			Enabled:   true,
			Values: []ruleengine.FeatureValue{
				{ID: uuid.New().String(), Value: "差旅费", Label: "差旅费", SortOrder: 1, Enabled: true},
				{ID: uuid.New().String(), Value: "办公费", Label: "办公费", SortOrder: 2, Enabled: true},
				{ID: uuid.New().String(), Value: "招待费", Label: "招待费", SortOrder: 3, Enabled: true},
				{ID: uuid.New().String(), Value: "日常费用", Label: "日常费用", SortOrder: 4, Enabled: true},
			},
		},
		{
			ID:        uuid.New().String(),
			Name:      "发票子分类",
			Code:      "invoice_subcategory",
			Type:      "string",
			ValueType: "single",
			Category:  "invoice",
			Enabled:   true,
			Values: []ruleengine.FeatureValue{
				{ID: uuid.New().String(), Value: "餐饮费", Label: "餐饮费", SortOrder: 1, Enabled: true},
				{ID: uuid.New().String(), Value: "办公用品", Label: "办公用品", SortOrder: 2, Enabled: true},
				{ID: uuid.New().String(), Value: "交通费", Label: "交通费", SortOrder: 3, Enabled: true},
				{ID: uuid.New().String(), Value: "住宿费", Label: "住宿费", SortOrder: 4, Enabled: true},
			},
		},
		{
			ID:        uuid.New().String(),
			Name:      "发票金额",
			Code:      "invoice_amount",
			Type:      "number",
			ValueType: "single",
			Category:  "invoice",
			Enabled:   true,
			Values:    []ruleengine.FeatureValue{},
		},
		{
			ID:        uuid.New().String(),
			Name:      "报销单金额",
			Code:      "reimbursement_amount",
			Type:      "number",
			ValueType: "single",
			Category:  "reimbursement",
			Enabled:   true,
			Values:    []ruleengine.FeatureValue{},
		},
	}

	for _, feature := range features {
		err := repo.CreateFeature(ctx, feature)
		if err != nil {
			return fmt.Errorf("创建特征 %s 失败: %w", feature.Code, err)
		}
		fmt.Printf("创建特征: %s - %s (ID: %s)\n", feature.Code, feature.Name, feature.ID)
	}

	return nil
}

func migrateRules(ctx context.Context, repo *mysql.RuleEngineRepository) error {
	features, _, err := repo.ListFeatures(ctx, &ruleengine.FeatureFilter{
		Page: 1,
		Size: 100,
	})
	if err != nil {
		return fmt.Errorf("获取特征列表失败: %w", err)
	}

	featureMap := make(map[string]string)
	for _, f := range features {
		featureMap[f.Code] = f.ID
	}

	rules := []*ruleengine.Rule{
		{
			ID:          uuid.New().String(),
			Name:        "日常费用单次限额500元",
			Description: "日常费用单次不得超过500元，超出部分需特殊审批",
			Priority:    90,
			Enabled:     true,
			Conditions: []ruleengine.Condition{
				{
					ID:        uuid.New().String(),
					FeatureID: featureMap["invoice_category"],
					Operator:  "eq",
					Value:     "日常费用",
					LogicOp:   "and",
					SortOrder: 0,
				},
				{
					ID:        uuid.New().String(),
					FeatureID: featureMap["invoice_amount"],
					Operator:  "gt",
					Value:     "500",
					LogicOp:   "and",
					SortOrder: 1,
				},
			},
			Decision: ruleengine.Decision{
				ID:     uuid.New().String(),
				Type:   "reject",
				Reason: "日常费用单次不得超过500元，超出部分需特殊审批",
			},
		},
		{
			ID:          uuid.New().String(),
			Name:        "交通费单次限额200元",
			Description: "交通费单次不得超过200元，超出部分需特殊审批",
			Priority:    89,
			Enabled:     true,
			Conditions: []ruleengine.Condition{
				{
					ID:        uuid.New().String(),
					FeatureID: featureMap["invoice_category"],
					Operator:  "eq",
					Value:     "差旅费",
					LogicOp:   "and",
					SortOrder: 0,
				},
				{
					ID:        uuid.New().String(),
					FeatureID: featureMap["invoice_subcategory"],
					Operator:  "eq",
					Value:     "交通费",
					LogicOp:   "and",
					SortOrder: 1,
				},
				{
					ID:        uuid.New().String(),
					FeatureID: featureMap["invoice_amount"],
					Operator:  "gt",
					Value:     "200",
					LogicOp:   "and",
					SortOrder: 2,
				},
			},
			Decision: ruleengine.Decision{
				ID:     uuid.New().String(),
				Type:   "reject",
				Reason: "交通费单次不得超过200元，超出部分需特殊审批",
			},
		},
		{
			ID:          uuid.New().String(),
			Name:        "餐饮费单次限额300元",
			Description: "餐饮费单次不得超过300元，超出部分需特殊审批",
			Priority:    88,
			Enabled:     true,
			Conditions: []ruleengine.Condition{
				{
					ID:        uuid.New().String(),
					FeatureID: featureMap["invoice_category"],
					Operator:  "eq",
					Value:     "差旅费",
					LogicOp:   "and",
					SortOrder: 0,
				},
				{
					ID:        uuid.New().String(),
					FeatureID: featureMap["invoice_subcategory"],
					Operator:  "eq",
					Value:     "餐饮费",
					LogicOp:   "and",
					SortOrder: 1,
				},
				{
					ID:        uuid.New().String(),
					FeatureID: featureMap["invoice_amount"],
					Operator:  "gt",
					Value:     "300",
					LogicOp:   "and",
					SortOrder: 2,
				},
			},
			Decision: ruleengine.Decision{
				ID:     uuid.New().String(),
				Type:   "reject",
				Reason: "餐饮费单次不得超过300元，超出部分需特殊审批",
			},
		},
		{
			ID:          uuid.New().String(),
			Name:        "办公用品单次限额1000元",
			Description: "办公用品单次不得超过1000元，超出部分需特殊审批",
			Priority:    87,
			Enabled:     true,
			Conditions: []ruleengine.Condition{
				{
					ID:        uuid.New().String(),
					FeatureID: featureMap["invoice_category"],
					Operator:  "eq",
					Value:     "办公费",
					LogicOp:   "and",
					SortOrder: 0,
				},
				{
					ID:        uuid.New().String(),
					FeatureID: featureMap["invoice_subcategory"],
					Operator:  "eq",
					Value:     "办公用品",
					LogicOp:   "and",
					SortOrder: 1,
				},
				{
					ID:        uuid.New().String(),
					FeatureID: featureMap["invoice_amount"],
					Operator:  "gt",
					Value:     "1000",
					LogicOp:   "and",
					SortOrder: 2,
				},
			},
			Decision: ruleengine.Decision{
				ID:     uuid.New().String(),
				Type:   "reject",
				Reason: "办公用品单次不得超过1000元，超出部分需特殊审批",
			},
		},
		{
			ID:          uuid.New().String(),
			Name:        "招待费单次限额2000元",
			Description: "招待费单次不得超过2000元，超出部分需特殊审批",
			Priority:    86,
			Enabled:     true,
			Conditions: []ruleengine.Condition{
				{
					ID:        uuid.New().String(),
					FeatureID: featureMap["invoice_category"],
					Operator:  "eq",
					Value:     "招待费",
					LogicOp:   "and",
					SortOrder: 0,
				},
				{
					ID:        uuid.New().String(),
					FeatureID: featureMap["invoice_amount"],
					Operator:  "gt",
					Value:     "2000",
					LogicOp:   "and",
					SortOrder: 1,
				},
			},
			Decision: ruleengine.Decision{
				ID:     uuid.New().String(),
				Type:   "reject",
				Reason: "招待费单次不得超过2000元，超出部分需特殊审批",
			},
		},
		{
			ID:          uuid.New().String(),
			Name:        "大额发票审批规则",
			Description: "单张发票金额超过5000元需要特殊审批",
			Priority:    80,
			Enabled:     true,
			Conditions: []ruleengine.Condition{
				{
					ID:        uuid.New().String(),
					FeatureID: featureMap["invoice_amount"],
					Operator:  "gt",
					Value:     "5000",
					LogicOp:   "and",
					SortOrder: 0,
				},
			},
			Decision: ruleengine.Decision{
				ID:     uuid.New().String(),
				Type:   "reject",
				Reason: "单张发票金额超过5000元，需要特殊审批",
			},
		},
		{
			ID:          uuid.New().String(),
			Name:        "超大额发票审批规则",
			Description: "单张发票金额超过10000元需要更高级别审批",
			Priority:    79,
			Enabled:     true,
			Conditions: []ruleengine.Condition{
				{
					ID:        uuid.New().String(),
					FeatureID: featureMap["invoice_amount"],
					Operator:  "gt",
					Value:     "10000",
					LogicOp:   "and",
					SortOrder: 0,
				},
			},
			Decision: ruleengine.Decision{
				ID:     uuid.New().String(),
				Type:   "reject",
				Reason: "单张发票金额超过10000元，需要更高级别审批",
			},
		},
	}

	for _, rule := range rules {
		err := repo.CreateRule(ctx, rule)
		if err != nil {
			return fmt.Errorf("创建规则 %s 失败: %w", rule.Name, err)
		}
		fmt.Printf("创建规则: %s (ID: %s)\n", rule.Name, rule.ID)
	}

	return nil
}
