package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"tunnel-server/internal/eventbus"
	"tunnel-server/internal/model"
	"tunnel-server/internal/pkg/validator"
)

// Create 校验隧道已联通、域名已登记且当前管理员有权限，然后写入并应用 nginx。
func (s *DomainService) Create(ctx context.Context, adminID, tunnelID uint, in CreateInputDomain) (*model.TunnelDomain, error) {
	tun, err := s.tunnels.FindByID(ctx, tunnelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if !tun.HasBeenOnline() {
		return nil, ErrTunnelOffline
	}
	host, err := validator.ValidateDomain(in.Domain)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrBadInput, err.Error())
	}
	if err := s.checkDomainAllowed(ctx, adminID, host); err != nil {
		return nil, err
	}
	if _, err := s.domains.FindByDomain(ctx, host); err == nil {
		return nil, ErrDomainTaken
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	row := &model.TunnelDomain{
		TunnelID:     tunnelID,
		Domain:       host,
		HTTPSEnabled: in.HTTPSEnabled,
		CertStatus:   model.CertStatusNone,
		NginxStatus:  model.NginxStatusPending,
	}
	if in.HTTPSEnabled {
		row.CertStatus = model.CertStatusPending
	}
	if err := s.domains.Create(ctx, row); err != nil {
		return nil, err
	}
	if err := s.applyNginx(ctx, row, false); err != nil {
		return row, err
	}
	s.afterChange(ctx, tunnelID)
	if in.HTTPSEnabled {
		go s.issueCert(row.ID)
	}
	return row, nil
}

// Delete 移除 nginx 配置与数据库记录。
func (s *DomainService) Delete(ctx context.Context, tunnelID, domainID uint) error {
	d, err := s.Get(ctx, tunnelID, domainID)
	if err != nil {
		return err
	}
	if err := s.nginx.Remove(d.Domain); err != nil {
		return err
	}
	if err := s.domains.Delete(ctx, d.ID); err != nil {
		return err
	}
	s.afterChange(ctx, tunnelID)
	return nil
}

func (s *DomainService) applyNginx(ctx context.Context, d *model.TunnelDomain, certReady bool) error {
	_, err := s.nginx.Apply(d.Domain, d.HTTPSEnabled, certReady)
	if err != nil {
		d.NginxStatus = model.NginxStatusFailed
		d.LastError = clip(err.Error(), 1024)
		_ = s.domains.Update(ctx, d)
		s.publish(d)
		return fmt.Errorf("%w: 应用 nginx 配置失败: %s", ErrOperation, err.Error())
	}
	if s.nginxOn {
		d.NginxStatus = model.NginxStatusApplied
	} else {
		d.NginxStatus = model.NginxStatusPreview
	}
	d.LastError = ""
	if err := s.domains.Update(ctx, d); err != nil {
		return err
	}
	s.publish(d)
	return nil
}

func (s *DomainService) afterChange(ctx context.Context, tunnelID uint) {
	if s.reload != nil {
		_ = s.reload.Reload(ctx)
	}
	if s.notify == nil {
		return
	}
	list, err := s.domains.ListByTunnel(ctx, tunnelID)
	if err != nil {
		return
	}
	names := make([]string, 0, len(list))
	for _, d := range list {
		names = append(names, d.Domain)
	}
	s.notify.NotifyConfig(tunnelID, names)
}

func (s *DomainService) publish(d *model.TunnelDomain) {
	if s.bus == nil {
		return
	}
	s.bus.Publish(eventbus.Event{Type: eventbus.EventDomainUpdated, Payload: d})
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func (s *DomainService) issueCert(domainID uint) {
	ctx := context.Background()
	d, err := s.domains.FindByID(ctx, domainID)
	if err != nil {
		return
	}
	if s.issue == nil || !s.certOn {
		d.CertStatus = model.CertStatusFailed
		d.LastError = "未启用 certbot，无法申请证书"
		_ = s.domains.Update(ctx, d)
		s.publish(d)
		return
	}
	expire, err := s.issue(ctx, d.Domain)
	if err != nil {
		d.CertStatus = model.CertStatusFailed
		d.LastError = clip(err.Error(), 1024)
		_ = s.domains.Update(ctx, d)
		s.publish(d)
		return
	}
	d.CertExpireAt = &expire
	d.CertStatus = model.CertStatusIssued
	d.LastError = ""
	_ = s.domains.Update(ctx, d)
	_ = s.applyNginx(ctx, d, true)
}
