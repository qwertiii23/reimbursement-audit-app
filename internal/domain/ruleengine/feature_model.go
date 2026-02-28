package ruleengine

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type FeatureConfig map[string]interface{}

func (fc FeatureConfig) Value() (driver.Value, error) {
	if len(fc) == 0 {
		return "{}", nil
	}
	return json.Marshal(fc)
}

func (fc *FeatureConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), fc)
	}
	return json.Unmarshal(bytes, fc)
}

type Feature struct {
	ID             string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name           string         `json:"name" gorm:"type:varchar(100);not null;uniqueIndex"`
	Code           string         `json:"code" gorm:"type:varchar(50);not null;uniqueIndex"`
	Description    string         `json:"description" gorm:"type:text"`
	Type           string         `json:"type" gorm:"type:varchar(20);not null"`
	ValueType      string         `json:"value_type" gorm:"type:varchar(20);not null"`
	Category       string         `json:"category" gorm:"type:varchar(50);index"`
	Enabled        bool           `json:"enabled" gorm:"default:true"`
	FunctionName   string         `json:"function_name" gorm:"type:varchar(100)"`
	FunctionConfig FeatureConfig  `json:"function_config" gorm:"type:json"`
	Values         []FeatureValue `json:"values" gorm:"foreignKey:FeatureID"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

type FeatureValue struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	FeatureID string    `json:"feature_id" gorm:"type:varchar(36);not null;index"`
	Value     string    `json:"value" gorm:"type:varchar(255);not null"`
	Label     string    `json:"label" gorm:"type:varchar(255);not null"`
	SortOrder int       `json:"sort_order" gorm:"default:0"`
	Enabled   bool      `json:"enabled" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type FeatureFilter struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	Category string `json:"category"`
	Type     string `json:"type"`
	Enabled  *bool  `json:"enabled"`
	Page     int    `json:"page"`
	Size     int    `json:"size"`
}

type FeatureValueFilter struct {
	FeatureID string `json:"feature_id"`
	Enabled   *bool  `json:"enabled"`
}

const (
	FeatureTypeString  = "string"
	FeatureTypeNumber  = "number"
	FeatureTypeBoolean = "boolean"
	FeatureTypeDate    = "date"
)

const (
	FeatureValueTypeSingle = "single"
	FeatureValueTypeList   = "list"
)

const (
	FeatureCategoryAmount        = "amount"
	FeatureCategoryInvoice       = "invoice"
	FeatureCategoryReimbursement = "reimbursement"
	FeatureCategoryUser          = "user"
	FeatureCategoryTime          = "time"
	FeatureCategoryOther         = "other"
)
