package model

import "time"

// 证书与 nginx 状态枚举。用字符串而非 iota，方便后台直接展示与跨语言对齐。
const (
	CertStatusNone    = "none"
	CertStatusPending = "pending"
	CertStatusIssued  = "issued"
	CertStatusFailed  = "failed"

	NginxStatusPreview = "preview"
	NginxStatusPending = "pending"
	NginxStatusApplied = "applied"
	NginxStatusFailed  = "failed"
)

// TunnelDomain 绑定到穿透记录的公网域名。
// domain 全局唯一，避免两条记录抢同一个 nginx server_name。
type TunnelDomain struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	TunnelID     uint       `json:"tunnel_id" gorm:"index;not null"`
	Domain       string     `json:"domain" gorm:"size:253;uniqueIndex;not null"`
	HTTPSEnabled bool       `json:"https_enabled"`
	CertStatus   string     `json:"cert_status" gorm:"size:16;not null;default:none"`
	CertExpireAt *time.Time `json:"cert_expire_at"`
	NginxStatus  string     `json:"nginx_status" gorm:"size:16;not null;default:pending"`
	LastError    string     `json:"last_error" gorm:"size:1024"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// TableName 指定表名。
func (TunnelDomain) TableName() string {
	return "tunnel_domains"
}

// CertReady 表示已拿到证书，可以切 443 配置。
func (d TunnelDomain) CertReady() bool {
	return d.HTTPSEnabled && d.CertStatus == CertStatusIssued
}
