package model

import "time"

// Domain 统一域名池中的一条可穿透域名。
// 域名统一为 <前缀>.<主域名>，由管理员自行创建；只能绑定到创建者本人的穿透记录。
type Domain struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	Domain string `json:"domain" gorm:"size:253;uniqueIndex;not null"`
	Remark string `json:"remark" gorm:"size:255"`
	// CreatedBy 创建该域名的管理员 ID。普通管理员只能看到、使用自己创建的域名。
	// 升级加列前的历史域名为 0：视为无归属，仅超级管理员可见可用。
	CreatedBy uint      `json:"created_by" gorm:"not null;default:0;index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (Domain) TableName() string {
	return "domains"
}
