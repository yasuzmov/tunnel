package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
)

type domainRepo struct {
	db *gorm.DB
}

// NewDomainRepo 创建域名仓储。
func NewDomainRepo(db *gorm.DB) DomainRepository {
	return &domainRepo{db: db}
}

func (r *domainRepo) Create(ctx context.Context, d *model.TunnelDomain) error {
	if err := r.db.WithContext(ctx).Create(d).Error; err != nil {
		return fmt.Errorf("创建域名: %w", err)
	}
	return nil
}

func (r *domainRepo) Update(ctx context.Context, d *model.TunnelDomain) error {
	if err := r.db.WithContext(ctx).Save(d).Error; err != nil {
		return fmt.Errorf("更新域名: %w", err)
	}
	return nil
}

func (r *domainRepo) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.TunnelDomain{}, id).Error; err != nil {
		return fmt.Errorf("删除域名: %w", err)
	}
	return nil
}

func (r *domainRepo) FindByID(ctx context.Context, id uint) (*model.TunnelDomain, error) {
	var d model.TunnelDomain
	if err := r.db.WithContext(ctx).First(&d, id).Error; err != nil {
		return nil, wrapFind("域名", err)
	}
	return &d, nil
}

func (r *domainRepo) FindByDomain(ctx context.Context, domain string) (*model.TunnelDomain, error) {
	var d model.TunnelDomain
	if err := r.db.WithContext(ctx).Where("domain = ?", domain).First(&d).Error; err != nil {
		return nil, wrapFind("域名", err)
	}
	return &d, nil
}

func (r *domainRepo) ListByTunnel(ctx context.Context, tunnelID uint) ([]model.TunnelDomain, error) {
	var list []model.TunnelDomain
	err := r.db.WithContext(ctx).Where("tunnel_id = ?", tunnelID).Order("id DESC").Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("列出域名: %w", err)
	}
	return list, nil
}

func (r *domainRepo) ListAll(ctx context.Context) ([]model.TunnelDomain, error) {
	var list []model.TunnelDomain
	if err := r.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, fmt.Errorf("列出全部域名: %w", err)
	}
	return list, nil
}

func (r *domainRepo) DeleteByTunnel(ctx context.Context, tunnelID uint) error {
	err := r.db.WithContext(ctx).Where("tunnel_id = ?", tunnelID).Delete(&model.TunnelDomain{}).Error
	if err != nil {
		return fmt.Errorf("按隧道删除域名: %w", err)
	}
	return nil
}
