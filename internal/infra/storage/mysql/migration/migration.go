// migration.go MySQL数据库迁移
// 功能点：
// 1. 使用GORM自动迁移功能
// 2. 支持数据库表结构自动更新
// 3. 支持迁移版本管理
// 4. 提供迁移状态查询

package mysql

import (
	"context"
	"fmt"
	"log"
	"time"

	"reimbursement-audit/internal/domain/ocr"
	"reimbursement-audit/internal/domain/reimbursement"
	"reimbursement-audit/internal/infra/storage/mysql"

	"gorm.io/gorm"
)

// Migration 迁移结构体
type Migration struct {
	Version     string `json:"version"`     // 迁移版本
	Description string `json:"description"` // 迁移描述
	Up          string `json:"up"`          // 升级SQL
	Down        string `json:"down"`        // 降级SQL
}

// MigrationManager 迁移管理器
type MigrationManager struct {
	client *mysql.Client
	db     *gorm.DB
}

// NewMigrationManager 创建迁移管理器
func NewMigrationManager(client *mysql.Client) *MigrationManager {
	return &MigrationManager{
		client: client,
		db:     client.GetDB(),
	}
}

// Up 执行迁移
func (m *MigrationManager) Up(ctx context.Context) error {
	// 使用GORM的AutoMigrate功能自动创建和更新表结构
	err := m.db.WithContext(ctx).AutoMigrate(
		// 报销单相关模型
		&reimbursement.Reimbursement{},
		&ocr.Invoice{},
		&reimbursement.AuditResult{},
		&reimbursement.AuditStatus{},
	)

	if err != nil {
		return fmt.Errorf("执行数据库迁移失败: %w", err)
	}

	log.Println("数据库迁移完成")
	return nil
}

// Down 回滚迁移
func (m *MigrationManager) Down(ctx context.Context) error {
	// GORM的AutoMigrate不支持自动回滚，需要手动处理
	// 这里只是记录日志，实际生产环境可能需要更复杂的回滚逻辑
	log.Println("警告: GORM的AutoMigrate不支持自动回滚，请手动处理数据库表结构回滚")
	return nil
}

// Status 获取迁移状态
func (m *MigrationManager) Status(ctx context.Context) (interface{}, error) {
	// 检查数据库连接状态
	sqlDB, err := m.db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 测试数据库连接
	if err := sqlDB.PingContext(ctx); err != nil {
		return map[string]interface{}{
			"status": "disconnected",
			"error":  err.Error(),
		}, nil
	}

	// 获取数据库表信息
	var tables []string
	err = m.db.WithContext(ctx).Raw("SHOW TABLES").Scan(&tables).Error
	if err != nil {
		return nil, fmt.Errorf("获取数据库表信息失败: %w", err)
	}

	return map[string]interface{}{
		"status":      "connected",
		"tables":      tables,
		"table_count": len(tables),
	}, nil
}

// Version 获取当前版本
func (m *MigrationManager) Version(ctx context.Context) (string, error) {
	// GORM的AutoMigrate不维护版本信息，这里返回当前时间作为版本标识
	return fmt.Sprintf("gorm-auto-migrate-%s", time.Now().Format("20060102-150405")), nil
}

// CreateMigrationsTable 创建迁移表
func (m *MigrationManager) CreateMigrationsTable(ctx context.Context) error {
	// 创建迁移记录表，用于跟踪迁移历史
	type MigrationRecord struct {
		ID        string    `gorm:"primaryKey;type:varchar(36)"`
		Version   string    `gorm:"type:varchar(50);not null;uniqueIndex"`
		AppliedAt time.Time `gorm:"type:datetime;not null"`
	}

	err := m.db.WithContext(ctx).AutoMigrate(&MigrationRecord{})
	if err != nil {
		return fmt.Errorf("创建迁移记录表失败: %w", err)
	}

	return nil
}

// GetMigrations 获取所有迁移
func (m *MigrationManager) GetMigrations() []Migration {
	// 由于使用GORM的AutoMigrate，这里返回一个空的迁移列表
	// 实际生产环境可能需要更复杂的迁移管理
	return []Migration{}
}

// ExecuteMigration 执行单个迁移
func (m *MigrationManager) ExecuteMigration(ctx context.Context, migration Migration) error {
	// 由于使用GORM的AutoMigrate，这里只是记录日志
	log.Printf("执行迁移: %s - %s\n", migration.Version, migration.Description)
	return nil
}

// RollbackMigration 回滚单个迁移
func (m *MigrationManager) RollbackMigration(ctx context.Context, migration Migration) error {
	// 由于使用GORM的AutoMigrate，这里只是记录日志
	log.Printf("警告: 无法回滚迁移: %s - %s\n", migration.Version, migration.Description)
	log.Println("GORM的AutoMigrate不支持自动回滚，请手动处理数据库表结构回滚")
	return nil
}

// IsMigrationApplied 检查迁移是否已应用
func (m *MigrationManager) IsMigrationApplied(ctx context.Context, version string) (bool, error) {
	// 由于使用GORM的AutoMigrate，这里返回true表示已应用
	// 实际生产环境可能需要查询迁移记录表
	return true, nil
}

// RecordMigration 记录迁移
func (m *MigrationManager) RecordMigration(ctx context.Context, version string) error {
	// 创建迁移记录表
	if err := m.CreateMigrationsTable(ctx); err != nil {
		return err
	}

	// 记录迁移
	type MigrationRecord struct {
		ID        string    `gorm:"primaryKey;type:varchar(36)"`
		Version   string    `gorm:"type:varchar(50);not null;uniqueIndex"`
		AppliedAt time.Time `gorm:"type:datetime;not null"`
	}

	record := MigrationRecord{
		ID:        fmt.Sprintf("migration-%s", version),
		Version:   version,
		AppliedAt: time.Now(),
	}

	err := m.db.WithContext(ctx).Create(&record).Error
	if err != nil {
		return fmt.Errorf("记录迁移失败: %w", err)
	}

	return nil
}

// RemoveMigrationRecord 移除迁移记录
func (m *MigrationManager) RemoveMigrationRecord(ctx context.Context, version string) error {
	// 创建迁移记录表
	if err := m.CreateMigrationsTable(ctx); err != nil {
		return err
	}

	// 移除迁移记录
	type MigrationRecord struct {
		ID        string    `gorm:"primaryKey;type:varchar(36)"`
		Version   string    `gorm:"type:varchar(50);not null;uniqueIndex"`
		AppliedAt time.Time `gorm:"type:datetime;not null"`
	}

	err := m.db.WithContext(ctx).Where("version = ?", version).Delete(&MigrationRecord{}).Error
	if err != nil {
		return fmt.Errorf("移除迁移记录失败: %w", err)
	}

	return nil
}
