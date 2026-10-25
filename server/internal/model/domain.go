package model

import "time"

// Domain 统一域名池中的一条可穿透域名。
// 只有在此表中登记、且当前管理员被授权的域名才允许绑定到穿透记录。
type Domain struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Domain    string    `json:"domain" gorm:"size:253;uniqueIndex;not null"`
	Remark    string    `json:"remark" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (Domain) TableName() string {
	return "domains"
}

// AdminDomain 管理员与可用域名的授权关系。
// 超级管理员不需要记录，默认拥有全部域名。
type AdminDomain struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	AdminID  uint `json:"admin_id" gorm:"uniqueIndex:uk_admin_domain;not null"`
	DomainID uint `json:"domain_id" gorm:"uniqueIndex:uk_admin_domain;not null"`
}

// TableName 指定表名。
func (AdminDomain) TableName() string {
	return "admin_domains"
}
