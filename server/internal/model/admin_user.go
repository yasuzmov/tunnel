// Package model 定义服务端持久化实体。
// 字段与表结构一一对应，不在此包编写业务规则。
package model

import "time"

// 管理员角色。超级管理员可管理管理员、域名池与授权；普通管理员只能使用被授权的域名。
const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
)

// AdminUser 管理后台账号。密码只存 bcrypt 哈希，禁止回写明文。
// Role 的默认值取 super_admin：升级旧库时已有账号（含 bootstrap）自动保留最高权限。
type AdminUser struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"size:64;uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"size:128;not null"`
	Role         string    `json:"role" gorm:"size:32;not null;default:super_admin"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// IsSuperAdmin 是否拥有全部域名与系统管理权限。
func (u AdminUser) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// TableName 指定表名。
func (AdminUser) TableName() string {
	return "admin_users"
}
