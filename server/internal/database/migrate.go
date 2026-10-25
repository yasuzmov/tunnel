package database

import (
	"fmt"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
)

// AutoMigrate 同步全部业务表。只增不改破坏性字段，生产变更需另行评审。
func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&model.AdminUser{},
		&model.Domain{},
		&model.AdminDomain{},
		&model.Tunnel{},
		&model.TunnelDomain{},
		&model.ClientSession{},
		&model.RequestLog{},
		&model.RequestLogBody{},
	)
	if err != nil {
		return fmt.Errorf("自动迁移: %w", err)
	}
	return nil
}
