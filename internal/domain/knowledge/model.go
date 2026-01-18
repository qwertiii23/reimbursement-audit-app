package knowledge

import (
	"time"
)

type KnowledgeFile struct {
	ID            string    `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`
	FileName      string    `json:"file_name" gorm:"type:varchar(255);not null;column:file_name"`
	FilePath      string    `json:"file_path" gorm:"type:varchar(500);not null;column:file_path"`
	FileType      string    `json:"file_type" gorm:"type:varchar(50);not null;column:file_type"`
	FileSize      int64     `json:"file_size" gorm:"type:bigint;column:file_size"`
	Category      string    `json:"category" gorm:"type:varchar(100);column:category"`
	Description   string    `json:"description" gorm:"type:text;column:description"`
	UploadedBy    string    `json:"uploaded_by" gorm:"type:varchar(36);column:uploaded_by"`
	UploaderName  string    `json:"uploader_name" gorm:"type:varchar(100);column:uploader_name"`
	DownloadCount int       `json:"download_count" gorm:"type:int;default:0;column:download_count"`
	Status        string    `json:"status" gorm:"type:varchar(20);default:'active';column:status"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type KnowledgeFileFilter struct {
	Category string `json:"category"`
	Status   string `json:"status"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}
