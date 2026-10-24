package model

import (
	"time"

	"gorm.io/gorm"
)

// Tunnel 一条内网穿透记录。Key 是客户端接入凭证，重置后旧连接立即失效。
type Tunnel struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name" gorm:"size:128;not null"`
	Remark       string         `json:"remark" gorm:"size:512"`
	Key          string         `json:"key" gorm:"size:64;uniqueIndex;not null"`
	LastOnlineAt *time.Time     `json:"last_online_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	Domains      []TunnelDomain `json:"domains,omitempty" gorm:"foreignKey:TunnelID"`
}

// TableName 指定表名。
func (Tunnel) TableName() string {
	return "tunnels"
}

// HasBeenOnline 表示至少成功联通过一次，此后才允许绑定域名。
func (t Tunnel) HasBeenOnline() bool {
	return t.LastOnlineAt != nil
}
