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
	// LogEnabled 关闭后公网请求仍转发，但不写 request_logs。
	// 默认开启：升级加列时已有记录继续记日志，避免静默停记。
	LogEnabled   bool           `json:"log_enabled" gorm:"not null;default:1"`
	LastOnlineAt *time.Time     `json:"last_online_at"`
	// CreatedBy 创建该记录的管理员 ID。普通管理员只能看到并操作自己创建的记录。
	// 升级加列前的历史记录为 0：视为无归属，仅超级管理员可见。
	CreatedBy    uint           `json:"created_by" gorm:"not null;default:0;index"`
	// CreatorName 创建者登录名，仅超级管理员查询时由服务层填充，不落库。
	CreatorName  string         `json:"creator_name,omitempty" gorm:"-"`
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
