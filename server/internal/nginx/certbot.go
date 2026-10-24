package nginx

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"tunnel-server/internal/config"
	"tunnel-server/internal/pkg/validator"
)

// Certbot 封装 Let's Encrypt 申请。续期交给系统 certbot.timer，这里只负责首次签发。
type Certbot struct {
	cfg    config.CertbotConfig
	nginx  config.NginxConfig
	runner CommandRunner
}

// NewCertbot 创建证书申请器。
func NewCertbot(cfg config.CertbotConfig, nginx config.NginxConfig, runner CommandRunner) *Certbot {
	if runner == nil {
		runner = ExecRunner{}
	}
	return &Certbot{cfg: cfg, nginx: nginx, runner: runner}
}

// Issue 通过 webroot 申请证书并返回到期时间。
func (c *Certbot) Issue(ctx context.Context, domain string) (time.Time, error) {
	domain, err := validator.ValidateDomain(domain)
	if err != nil {
		return time.Time{}, err
	}
	if !c.cfg.Enabled {
		return time.Time{}, fmt.Errorf("未启用 certbot，无法申请证书")
	}
	if c.cfg.Email == "" {
		return time.Time{}, fmt.Errorf("certbot.email 未配置")
	}
	webroot := c.cfg.Webroot
	if webroot == "" {
		webroot = c.nginx.ACMEWebroot
	}
	if err := os.MkdirAll(filepath.Join(webroot), 0o755); err != nil {
		return time.Time{}, fmt.Errorf("创建 acme webroot: %w", err)
	}
	bin := c.cfg.Bin
	if bin == "" {
		bin = "certbot"
	}
	args := []string{
		"certonly", "--webroot",
		"-w", webroot,
		"-d", domain,
		"--non-interactive", "--agree-tos",
		"-m", c.cfg.Email,
		"--deploy-hook", "nginx -s reload",
	}
	if _, err := c.runner.Run(ctx, bin, args...); err != nil {
		return time.Time{}, fmt.Errorf("certbot 申请失败: %w", err)
	}
	return ReadCertExpiry(domain)
}

// ReadCertExpiry 解析 let's encrypt 证书的 NotAfter。
func ReadCertExpiry(domain string) (time.Time, error) {
	path := filepath.Join("/etc/letsencrypt/live", domain, "cert.pem")
	data, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}, fmt.Errorf("读取证书: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return time.Time{}, fmt.Errorf("证书 PEM 无效")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, fmt.Errorf("解析证书: %w", err)
	}
	return cert.NotAfter, nil
}
