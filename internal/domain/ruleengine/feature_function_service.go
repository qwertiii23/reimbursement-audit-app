package ruleengine

import (
	"context"
	"reimbursement-audit/internal/domain/featurefunction"
	"reimbursement-audit/internal/pkg/logger"
)

type FeatureFunctionService struct {
	registry *featurefunction.FunctionRegistry
	logger   logger.Logger
}

func NewFeatureFunctionService(
	registry *featurefunction.FunctionRegistry,
	logger logger.Logger,
) *FeatureFunctionService {
	return &FeatureFunctionService{
		registry: registry,
		logger:   logger,
	}
}

type FunctionSchemaResponse struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	ConfigSchema *featurefunction.ConfigSchema `json:"config_schema"`
}

func (s *FeatureFunctionService) ListFunctions(ctx context.Context) []*FunctionSchemaResponse {
	functions := s.registry.List()
	
	responses := make([]*FunctionSchemaResponse, 0, len(functions))
	for _, fn := range functions {
		responses = append(responses, &FunctionSchemaResponse{
			Name:        fn.GetName(),
			Description: fn.GetDescription(),
			ConfigSchema: fn.GetConfigSchema(),
		})
	}
	
	return responses
}

func (s *FeatureFunctionService) GetFunctionSchema(ctx context.Context, name string) (*FunctionSchemaResponse, error) {
	schema, err := s.registry.GetSchema(name)
	if err != nil {
		return nil, err
	}
	
	fn, _ := s.registry.Get(name)
	
	return &FunctionSchemaResponse{
		Name:        fn.GetName(),
		Description: fn.GetDescription(),
		ConfigSchema: schema,
	}, nil
}