// Package config 负责加载服务端 YAML 配置。
// 所有可调超时、路径、密钥都集中在这里，避免散落在业务代码中写死。
package config

import (
	"net"
	"net/url"
	"strings"
	"time"
)

// Config 是服务端运行时配置的根结构。
type Config struct {
	Admin     AdminConfig     `yaml:"admin"`
	Ingress   IngressConfig   `yaml:"ingress"`
	MySQL     MySQLConfig     `yaml:"mysql"`
	JWT       JWTConfig       `yaml:"jwt"`
	Bootstrap BootstrapConfig `yaml:"bootstrap"`
	Nginx     NginxConfig     `yaml:"nginx"`
	Certbot   CertbotConfig   `yaml:"certbot"`
	Log       LogConfig       `yaml:"log"`
	Tunnel    TunnelConfig    `yaml:"tunnel"`
}

// AdminConfig 管理 API 与后台页面监听地址。
type AdminConfig struct {
	Listen    string `yaml:"listen"`
	PublicURL string `yaml:"public_url"`
	// DomainSuffix 域名池统一后缀（DNS 中已配置 *.后缀 泛解析）。
	// 留空时取 public_url 的主机名，例如 https://fish.czypro.cn -> fish.czypro.cn。
	DomainSuffix string `yaml:"domain_suffix"`
}

// IngressConfig 公网流量入口，仅给 nginx 反代使用。
type IngressConfig struct {
	Listen string `yaml:"listen"`
}

// MySQLConfig 数据库连接。
type MySQLConfig struct {
	DSN string `yaml:"dsn"`
}

// JWTConfig 管理后台鉴权。
type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

// BootstrapConfig 首次启动时写入的初始管理员。
// 若库中已有管理员则不会覆盖密码，避免重启把运营改过的口令冲掉。
type BootstrapConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// NginxConfig 控制是否真正落盘 nginx 配置。
// enabled=false 时只预览模板，方便 Windows 本机联调。
type NginxConfig struct {
	Enabled           bool   `yaml:"enabled"`
	ConfDir           string `yaml:"conf_dir"`
	NginxBin          string `yaml:"nginx_bin"`
	ClientMaxBodySize string `yaml:"client_max_body_size"`
	ACMEWebroot       string `yaml:"acme_webroot"`
	IngressUpstream   string `yaml:"ingress_upstream"`
}

// CertbotConfig Let's Encrypt 申请参数。
type CertbotConfig struct {
	Enabled bool   `yaml:"enabled"`
	Bin     string `yaml:"bin"`
	Email   string `yaml:"email"`
	Webroot string `yaml:"webroot"`
}

// LogConfig 请求日志截断、批量写入与保留策略。
type LogConfig struct {
	BodyLimit       int `yaml:"body_limit"`
	RetainDays      int `yaml:"retain_days"`
	BatchSize       int `yaml:"batch_size"`
	FlushIntervalMS int `yaml:"flush_interval_ms"`
	ChannelSize     int `yaml:"channel_size"`
}

// TunnelConfig 隧道空闲与转发超时。
type TunnelConfig struct {
	IdleTimeoutSec     int `yaml:"idle_timeout_sec"`
	ResponseTimeoutSec int `yaml:"response_timeout_sec"`
}

// ExpireDuration 返回 JWT 有效期。
func (c JWTConfig) ExpireDuration() time.Duration {
	hours := c.ExpireHours
	if hours <= 0 {
		hours = 72
	}
	return time.Duration(hours) * time.Hour
}

// IdleTimeout 返回隧道空闲超时。
func (c TunnelConfig) IdleTimeout() time.Duration {
	sec := c.IdleTimeoutSec
	if sec <= 0 {
		sec = 45
	}
	return time.Duration(sec) * time.Second
}

// ResponseTimeout 返回等待客户端响应头的超时。
func (c TunnelConfig) ResponseTimeout() time.Duration {
	sec := c.ResponseTimeoutSec
	if sec <= 0 {
		sec = 60
	}
	return time.Duration(sec) * time.Second
}

// FlushInterval 返回日志批量刷盘间隔。
func (c LogConfig) FlushInterval() time.Duration {
	ms := c.FlushIntervalMS
	if ms <= 0 {
		ms = 1000
	}
	return time.Duration(ms) * time.Millisecond
}

// EffectiveBodyLimit 返回请求/响应体截断上限。
func (c LogConfig) EffectiveBodyLimit() int {
	if c.BodyLimit <= 0 {
		return 64 * 1024
	}
	return c.BodyLimit
}

// BaseDomain 返回域名池统一后缀：优先 domain_suffix，否则取 public_url 的主机名。
// 主机名为 IP、localhost 等不含点的名字时返回空串，表示未配置，域名池将拒绝新增。
func (c AdminConfig) BaseDomain() string {
	raw := strings.TrimSpace(c.DomainSuffix)
	if raw == "" {
		u, err := url.Parse(strings.TrimSpace(c.PublicURL))
		if err != nil {
			return ""
		}
		raw = u.Hostname()
	}
	host := strings.Trim(strings.ToLower(strings.TrimSpace(raw)), ".")
	host = strings.TrimPrefix(host, "*.")
	if host == "" || !strings.Contains(host, ".") || net.ParseIP(host) != nil {
		return ""
	}
	return host
}
