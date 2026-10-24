// Package model 定义服务端持久化实体。
// 字段与表结构一一对应，不在此包编写业务规则。
package model

import "time"

// AdminUser 管理后台账号。密码只存 bcrypt 哈希，禁止回写明文。
type AdminUser struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"size:64;uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"size:128;not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (AdminUser) TableName() string {
	return "admin_users"
}
