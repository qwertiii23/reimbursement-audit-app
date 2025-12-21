package validator

import (
	"fmt"
	"reflect"
	"strings"
)

// validatorImpl 验证器实现
type validatorImpl struct {
	rules     map[string][]Rule     // 字段验证规则
	validators map[string]Validator   // 嵌套验证器
	config    *Config               // 配置
}

// NewValidator 创建验证器实例
func NewValidator(config *Config) Validator {
	if config == nil {
		config = DefaultConfig()
	}
	
	return &validatorImpl{
		rules:     make(map[string][]Rule),
		validators: make(map[string]Validator),
		config:    config,
	}
}

// Validate 验证对象
func (v *validatorImpl) Validate(obj interface{}) error {
	if obj == nil {
		return fmt.Errorf("object to validate is nil")
	}
	
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("object to validate must be a struct or pointer to struct")
	}
	
	var errors ValidationErrors
	
	// 验证字段
	for field, rules := range v.rules {
		fieldValue := v.getFieldValue(obj, field)
		for _, rule := range rules {
			if err := rule.Validate(fieldValue); err != nil {
				errors = append(errors, ValidationError{
					Field:   field,
					Message: err.Error(),
					Value:   fieldValue,
				})
				break // 一个字段只返回第一个错误
			}
		}
	}
	
	// 验证嵌套对象
	for field, validator := range v.validators {
		fieldValue := v.getFieldValue(obj, field)
		if err := validator.Validate(fieldValue); err != nil {
			if validationErrors, ok := err.(ValidationErrors); ok {
				for _, validationError := range validationErrors {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("%s.%s", field, validationError.Field),
						Message: validationError.Message,
						Value:   validationError.Value,
					})
				}
			} else {
				errors = append(errors, ValidationError{
					Field:   field,
					Message: err.Error(),
					Value:   fieldValue,
				})
			}
		}
	}
	
	if len(errors) > 0 {
		return errors
	}
	
	return nil
}

// AddRule 添加验证规则
func (v *validatorImpl) AddRule(field string, rule Rule) Validator {
	if v.rules[field] == nil {
		v.rules[field] = make([]Rule, 0)
	}
	v.rules[field] = append(v.rules[field], rule)
	return v
}

// AddRules 添加多个验证规则
func (v *validatorImpl) AddRules(field string, rules []Rule) Validator {
	if v.rules[field] == nil {
		v.rules[field] = make([]Rule, 0)
	}
	v.rules[field] = append(v.rules[field], rules...)
	return v
}

// AddValidator 添加嵌套验证器
func (v *validatorImpl) AddValidator(field string, validator Validator) Validator {
	v.validators[field] = validator
	return v
}

// getFieldValue 获取字段值
func (v *validatorImpl) getFieldValue(obj interface{}, field string) interface{} {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	// 处理嵌套字段，如 "user.profile.name"
	fields := strings.Split(field, ".")
	for _, f := range fields {
		if val.Kind() == reflect.Struct {
			field := val.FieldByName(f)
			if !field.IsValid() {
				return nil
			}
			val = field
		} else if val.Kind() == reflect.Map {
			key := reflect.ValueOf(f)
			field := val.MapIndex(key)
			if !field.IsValid() {
				return nil
			}
			val = field
		} else {
			return nil
		}
	}
	
	return val.Interface()
}

// FieldValidator 字段验证器
type FieldValidator struct {
	field  string
	rules  []Rule
	parent Validator
}

// NewFieldValidator 创建字段验证器
func NewFieldValidator(field string, rules []Rule, parent Validator) *FieldValidator {
	return &FieldValidator{
		field:  field,
		rules:  rules,
		parent: parent,
	}
}

// Validate 验证字段
func (fv *FieldValidator) Validate(obj interface{}) error {
	value := GetFieldValue(obj, fv.field)
	for _, rule := range fv.rules {
		if err := rule.Validate(value); err != nil {
			return ValidationError{
				Field:   fv.field,
				Message: err.Error(),
				Value:   value,
			}
		}
	}
	return nil
}

// AddRule 添加验证规则
func (fv *FieldValidator) AddRule(field string, rule Rule) Validator {
	return fv.parent.AddRule(field, rule)
}

// AddRules 添加多个验证规则
func (fv *FieldValidator) AddRules(field string, rules []Rule) Validator {
	return fv.parent.AddRules(field, rules)
}

// AddValidator 添加嵌套验证器
func (fv *FieldValidator) AddValidator(field string, validator Validator) Validator {
	return fv.parent.AddValidator(field, validator)
}