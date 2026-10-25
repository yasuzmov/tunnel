package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
	"tunnel-server/internal/pkg/crypto"
	"tunnel-server/internal/repository"
)

// TunnelDisconnect 删除或重置 key 时断开在线客户端。
type TunnelDisconnect func(tunnelID uint, reason string)

// TunnelDeps 穿透记录服务依赖。
type TunnelDeps struct {
	Tunnels    repository.TunnelRepository
	Domains    repository.DomainRepository
	PublicURL  string
	Disconnect TunnelDisconnect
	RemoveConf func(domain string) error
	Reload     DomainReloader
}

// TunnelService 穿透记录的创建、更新、重置 key。
type TunnelService struct {
	tunnels    repository.TunnelRepository
	domains    repository.DomainRepository
	disconnect TunnelDisconnect
	removeConf func(domain string) error
	reload     DomainReloader
	publicURL  string
}

// NewTunnelService 创建穿透记录服务。
func NewTunnelService(d TunnelDeps) *TunnelService {
	return &TunnelService{
		tunnels:    d.Tunnels,
		domains:    d.Domains,
		disconnect: d.Disconnect,
		removeConf: d.RemoveConf,
		reload:     d.Reload,
		publicURL:  d.PublicURL,
	}
}

// CreateInput 新建穿透记录。
type CreateInput struct {
	Name   string `json:"name"`
	Remark string `json:"remark"`
}

// UpdateInput 更新展示字段，不改 key。
type UpdateInput struct {
	Name   string `json:"name"`
	Remark string `json:"remark"`
}

// SetLoggingInput 是否记录该穿透的转发日志。
type SetLoggingInput struct {
	Enabled bool `json:"enabled"`
}

// Create 生成随机 key 并落库。
func (s *TunnelService) Create(ctx context.Context, in CreateInput) (*model.Tunnel, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: 名称不能为空", ErrBadInput)
	}
	key, err := crypto.NewTunnelKey()
	if err != nil {
		return nil, err
	}
	t := &model.Tunnel{Name: name, Remark: strings.TrimSpace(in.Remark), Key: key, LogEnabled: true}
	if err := s.tunnels.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// SetLogging 切换该记录是否写入转发日志，并刷新入口缓存使开关立即生效。
// 缓存刷新失败时数据库已更新，返回 ErrOperation，再次切换会重试刷新。
func (s *TunnelService) SetLogging(ctx context.Context, id uint, enabled bool) (*model.Tunnel, error) {
	t, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if t.LogEnabled == enabled {
		return t, nil
	}
	t.LogEnabled = enabled
	if err := s.tunnels.Update(ctx, t); err != nil {
		return nil, err
	}
	if s.reload != nil {
		if err := s.reload.Reload(ctx); err != nil {
			return nil, fmt.Errorf("%w: 刷新入口缓存失败", ErrOperation)
		}
	}
	return t, nil
}

// Update 修改名称与备注。
func (s *TunnelService) Update(ctx context.Context, id uint, in UpdateInput) (*model.Tunnel, error) {
	t, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: 名称不能为空", ErrBadInput)
	}
	t.Name = name
	t.Remark = strings.TrimSpace(in.Remark)
	if err := s.tunnels.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// Get 读取单条记录。
func (s *TunnelService) Get(ctx context.Context, id uint) (*model.Tunnel, error) {
	t, err := s.tunnels.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

// List 分页列表。
func (s *TunnelService) List(ctx context.Context, page, pageSize int, keyword string) ([]model.Tunnel, int64, error) {
	return s.tunnels.List(ctx, page, pageSize, strings.TrimSpace(keyword))
}

// Delete 软删除记录、清理 nginx 配置并踢掉在线客户端。
func (s *TunnelService) Delete(ctx context.Context, id uint) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}
	if s.disconnect != nil {
		s.disconnect(id, model.DisconnectDeleted)
	}
	list, _ := s.domains.ListByTunnel(ctx, id)
	for _, d := range list {
		if s.removeConf != nil {
			_ = s.removeConf(d.Domain)
		}
	}
	_ = s.domains.DeleteByTunnel(ctx, id)
	if s.reload != nil {
		_ = s.reload.Reload(ctx)
	}
	return s.tunnels.Delete(ctx, id)
}

// ResetKey 轮换接入凭证，旧客户端会立即鉴权失败。
func (s *TunnelService) ResetKey(ctx context.Context, id uint) (*model.Tunnel, error) {
	t, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	key, err := crypto.NewTunnelKey()
	if err != nil {
		return nil, err
	}
	t.Key = key
	if err := s.tunnels.Update(ctx, t); err != nil {
		return nil, err
	}
	if s.disconnect != nil {
		s.disconnect(id, model.DisconnectResetKey)
	}
	return t, nil
}

// PublicURL 返回配置的管理端公网地址，用于拼接客户端接入 URL。
func (s *TunnelService) PublicURL() string {
	return strings.TrimRight(s.publicURL, "/")
}

// ConnectURL 客户端 WebSocket 接入地址。
func (s *TunnelService) ConnectURL() string {
	base := s.PublicURL()
	switch {
	case strings.HasPrefix(base, "https://"):
		return "wss://" + strings.TrimPrefix(base, "https://") + "/tunnel/ws"
	case strings.HasPrefix(base, "http://"):
		return "ws://" + strings.TrimPrefix(base, "http://") + "/tunnel/ws"
	default:
		return base + "/tunnel/ws"
	}
}
