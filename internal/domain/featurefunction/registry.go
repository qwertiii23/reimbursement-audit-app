package featurefunction

import (
	"context"
	"fmt"
	"sync"
)

type FeatureFunction interface {
	GetName() string
	GetDescription() string
	GetConfigSchema() *ConfigSchema
	Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error)
	Validate(config map[string]interface{}) error
}

type ConfigSchema struct {
	Fields []FieldConfig `json:"fields"`
}

type FieldConfig struct {
	Name        string `json:"name"`
	Type        string `json:"type"`        // string, number, boolean, select
	Label       string `json:"label"`
	Required    bool   `json:"required"`
	Default     interface{} `json:"default"`
	Options     []Option `json:"options,omitempty"` // for select type
	Description string `json:"description,omitempty"`
}

type Option struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

type FunctionInput struct {
	Config      map[string]interface{} `json:"config"`
	InvoiceData map[string]interface{} `json:"invoice_data"`
	InvoiceID   string               `json:"invoice_id"`
	Params      map[string]interface{} `json:"params"`
}

type FunctionOutput struct {
	Value       interface{} `json:"value"`
	Confidence  float64     `json:"confidence,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Error       string      `json:"error,omitempty"`
}

type FunctionRegistry struct {
	functions map[string]FeatureFunction
	mu        sync.RWMutex
}

func NewFunctionRegistry() *FunctionRegistry {
	return &FunctionRegistry{
		functions: make(map[string]FeatureFunction),
	}
}

func (r *FunctionRegistry) Register(fn FeatureFunction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := fn.GetName()
	if name == "" {
		return fmt.Errorf("function name cannot be empty")
	}

	if _, exists := r.functions[name]; exists {
		return fmt.Errorf("function %s already registered", name)
	}

	r.functions[name] = fn
	return nil
}

func (r *FunctionRegistry) Get(name string) (FeatureFunction, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fn, exists := r.functions[name]
	return fn, exists
}

func (r *FunctionRegistry) List() []FeatureFunction {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]FeatureFunction, 0, len(r.functions))
	for _, fn := range r.functions {
		list = append(list, fn)
	}
	return list
}

func (r *FunctionRegistry) GetSchema(name string) (*ConfigSchema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fn, exists := r.functions[name]
	if !exists {
		return nil, fmt.Errorf("function %s not found", name)
	}

	return fn.GetConfigSchema(), nil
}