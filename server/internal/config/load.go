package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load 从 YAML 文件读取配置，并补齐开发联调所需的默认值。
// 缺省值保证 Windows 本机不装 nginx 也能跑通隧道与日志。
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件 %s: %w", path, err)
	}
	cfg := defaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件 %s: %w", path, err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Admin: AdminConfig{
			Listen:    ":8080",
			PublicURL: "http://127.0.0.1:8080",
		},
		Ingress: IngressConfig{
			Listen: "127.0.0.1:8081",
		},
		JWT: JWTConfig{
			ExpireHours: 72,
		},
		Bootstrap: BootstrapConfig{
			Username: "admin",
			Password: "admin123",
		},
		Nginx: NginxConfig{
			Enabled:           false,
			ConfDir:           "/etc/nginx/conf.d/tunnel",
			NginxBin:          "nginx",
			ClientMaxBodySize: "50m",
			ACMEWebroot:       "/var/www/acme",
			IngressUpstream:   "127.0.0.1:8081",
		},
		Certbot: CertbotConfig{
			Enabled: false,
			Bin:     "certbot",
			Webroot: "/var/www/acme",
		},
		Log: LogConfig{
			BodyLimit:       65536,
			RetainDays:      30,
			BatchSize:       200,
			FlushIntervalMS: 1000,
			ChannelSize:     4096,
		},
		Tunnel: TunnelConfig{
			IdleTimeoutSec:     45,
			ResponseTimeoutSec: 60,
		},
	}
}

func (c *Config) validate() error {
	if c.MySQL.DSN == "" {
		return fmt.Errorf("mysql.dsn 不能为空")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret 不能为空")
	}
	if c.Admin.Listen == "" {
		return fmt.Errorf("admin.listen 不能为空")
	}
	if c.Ingress.Listen == "" {
		return fmt.Errorf("ingress.listen 不能为空")
	}
	return nil
}
