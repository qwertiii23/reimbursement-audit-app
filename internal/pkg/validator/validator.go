package validator

import (
	"fmt"
	"reflect"
	"strings"
)

// Rule 验证规则
type Rule struct {
	Name     string                 // 规则名称
	Validate func(interface{}) error // 验证函数
	Message  string                 // 错误消息
}

// Validator 验证器接口
type Validator interface {
	// Validate 验证对象
	Validate(obj interface{}) error
	// AddRule 添加验证规则
	AddRule(field string, rule Rule) Validator
	// AddRules 添加多个验证规则
	AddRules(field string, rules []Rule) Validator
	// AddValidator 添加嵌套验证器
	AddValidator(field string, validator Validator) Validator
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string // 字段名
	Message string // 错误消息
	Value   interface{} // 字段值
}

// Error 实现error接口
func (e ValidationError) Error() string {
	return fmt.Sprintf("field '%s' validation failed: %s", e.Field, e.Message)
}

// ValidationErrors 验证错误集合
type ValidationErrors []ValidationError

// Error 实现error接口
func (es ValidationErrors) Error() string {
	if len(es) == 0 {
		return ""
	}
	
	var messages []string
	for _, err := range es {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// Config 验证器配置
type Config struct {
	// TODO: 添加验证器配置字段
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		// TODO: 设置默认配置
	}
}

// ValidateStruct 验证结构体
func ValidateStruct(obj interface{}) error {
	// TODO: 实现结构体验证逻辑
	return nil
}

// ValidateField 验证字段
func ValidateField(field interface{}, rules []Rule) error {
	// TODO: 实现字段验证逻辑
	return nil
}

// Required 必填验证
func Required(value interface{}) error {
	// TODO: 实现必填验证逻辑
	return nil
}

// MinLength 最小长度验证
func MinLength(min int) func(interface{}) error {
	return func(value interface{}) error {
		// TODO: 实现最小长度验证逻辑
		return nil
	}
}

// MaxLength 最大长度验证
func MaxLength(max int) func(interface{}) error {
	return func(value interface{}) error {
		// TODO: 实现最大长度验证逻辑
		return nil
	}
}

// Range 范围验证
func Range(min, max int) func(interface{}) error {
	return func(value interface{}) error {
		// TODO: 实现范围验证逻辑
		return nil
	}
}

// Email 邮箱验证
func Email(value interface{}) error {
	// TODO: 实现邮箱验证逻辑
	return nil
}

// Phone 手机号验证
func Phone(value interface{}) error {
	// TODO: 实现手机号验证逻辑
	return nil
}

// IDCard 身份证验证
func IDCard(value interface{}) error {
	// TODO: 实现身份证验证逻辑
	return nil
}

// Regex 正则验证
func Regex(pattern string) func(interface{}) error {
	return func(value interface{}) error {
		// TODO: 实现正则验证逻辑
		return nil
	}
}

// Custom 自定义验证
func Custom(validate func(interface{}) bool, message string) func(interface{}) error {
	return func(value interface{}) error {
		if !validate(value) {
			return fmt.Errorf("%s", message)
		}
		return nil
	}
}

// GetFieldValue 获取字段值
func GetFieldValue(obj interface{}, field string) interface{} {
	// TODO: 实现获取字段值逻辑
	return nil
}

// SetFieldValue 设置字段值
func SetFieldValue(obj interface{}, field string, value interface{}) error {
	// TODO: 实现设置字段值逻辑
	return nil
}

// GetFieldType 获取字段类型
func GetFieldType(obj interface{}, field string) reflect.Type {
	// TODO: 实现获取字段类型逻辑
	return nil
}