package nginx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"tunnel-server/internal/config"
	"tunnel-server/internal/pkg/validator"
)

// Manager 把域名配置写入独立 conf 文件，经 nginx -t 成功后才 reload。
type Manager struct {
	cfg    config.NginxConfig
	runner CommandRunner
}

// NewManager 创建 nginx 管理器。runner 为 nil 时使用 ExecRunner。
func NewManager(cfg config.NginxConfig, runner CommandRunner) *Manager {
	if runner == nil {
		runner = ExecRunner{}
	}
	return &Manager{cfg: cfg, runner: runner}
}

// Preview 只渲染不落盘，供后台展示和 Windows 联调。
func (m *Manager) Preview(domain string, https, certReady bool) (string, error) {
	return Render(m.input(domain, https, certReady))
}

// Apply 写入配置并 reload。enabled=false 时返回预览文本，状态由调用方记为 preview。
func (m *Manager) Apply(domain string, https, certReady bool) (string, error) {
	conf, err := m.Preview(domain, https, certReady)
	if err != nil {
		return "", err
	}
	if !m.cfg.Enabled {
		return conf, nil
	}
	if err := m.writeAndReload(domain, conf); err != nil {
		return conf, err
	}
	return conf, nil
}

// Remove 删除域名对应的 conf 并 reload。文件不存在视为成功。
func (m *Manager) Remove(domain string) error {
	domain, err := validator.ValidateDomain(domain)
	if err != nil {
		return err
	}
	if !m.cfg.Enabled {
		return nil
	}
	path := m.confPath(domain)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除 nginx 配置: %w", err)
	}
	return m.reload(context.Background())
}

func (m *Manager) input(domain string, https, certReady bool) RenderInput {
	return RenderInput{
		Domain:            domain,
		Upstream:          m.cfg.IngressUpstream,
		ClientMaxBodySize: m.cfg.ClientMaxBodySize,
		ACMEWebroot:       m.cfg.ACMEWebroot,
		HTTPS:             https,
		CertReady:         certReady,
	}
}

func (m *Manager) confPath(domain string) string {
	return filepath.Join(m.cfg.ConfDir, domain+".conf")
}

func (m *Manager) writeAndReload(domain, conf string) error {
	if err := os.MkdirAll(m.cfg.ConfDir, 0o755); err != nil {
		return fmt.Errorf("创建 nginx 配置目录: %w", err)
	}
	path := m.confPath(domain)
	backup, err := backupFile(path)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(conf), 0o644); err != nil {
		return fmt.Errorf("写入 nginx 配置: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := m.runner.Run(ctx, m.bin(), "-t"); err != nil {
		_ = restoreFile(path, backup)
		return fmt.Errorf("nginx -t 失败: %w", err)
	}
	return m.reload(ctx)
}

func (m *Manager) reload(ctx context.Context) error {
	if _, err := m.runner.Run(ctx, m.bin(), "-s", "reload"); err != nil {
		return fmt.Errorf("nginx reload 失败: %w", err)
	}
	return nil
}

func (m *Manager) bin() string {
	if m.cfg.NginxBin == "" {
		return "nginx"
	}
	return m.cfg.NginxBin
}

func backupFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("读取旧 nginx 配置: %w", err)
	}
	return data, nil
}

func restoreFile(path string, backup []byte) error {
	if backup == nil {
		_ = os.Remove(path)
		return nil
	}
	return os.WriteFile(path, backup, 0o644)
}
