package database

import (
	"fmt"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
	"tunnel-server/internal/pkg/crypto"
)

// SeedAdmin 仅在管理员表为空时写入初始账号。
// 已有账号时绝不覆盖密码，避免重启冲掉运营改过的口令。
func SeedAdmin(db *gorm.DB, username, password string) error {
	if username == "" || password == "" {
		return nil
	}
	var count int64
	if err := db.Model(&model.AdminUser{}).Count(&count).Error; err != nil {
		return fmt.Errorf("统计管理员: %w", err)
	}
	if count > 0 {
		return nil
	}
	hashed, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}
	user := &model.AdminUser{Username: username, PasswordHash: hashed}
	if err := db.Create(user).Error; err != nil {
		return fmt.Errorf("写入初始管理员: %w", err)
	}
	return nil
}
