package database

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
	user := &model.AdminUser{Username: username, PasswordHash: hashed, Role: model.RoleSuperAdmin}
	if err := db.Create(user).Error; err != nil {
		return fmt.Errorf("写入初始管理员: %w", err)
	}
	return nil
}

// SeedDomainPool 把升级前已经绑定到穿透记录的域名补进统一域名池，
// 避免引入域名池后旧域名无法继续使用。重复执行不会产生重复行。
func SeedDomainPool(db *gorm.DB) error {
	var hosts []string
	if err := db.Model(&model.TunnelDomain{}).Distinct().Pluck("domain", &hosts).Error; err != nil {
		return fmt.Errorf("读取已有域名: %w", err)
	}
	if len(hosts) == 0 {
		return nil
	}
	rows := make([]model.Domain, 0, len(hosts))
	for _, host := range hosts {
		rows = append(rows, model.Domain{Domain: host, Remark: "升级时自动同步"})
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error; err != nil {
		return fmt.Errorf("同步域名池: %w", err)
	}
	return nil
}
