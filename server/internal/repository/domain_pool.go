package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
)

type domainPoolRepo struct {
	db *gorm.DB
}

// NewDomainPoolRepo 创建统一域名池仓储。
func NewDomainPoolRepo(db *gorm.DB) DomainPoolRepository {
	return &domainPoolRepo{db: db}
}

func (r *domainPoolRepo) Create(ctx context.Context, d *model.Domain) error {
	if err := r.db.WithContext(ctx).Create(d).Error; err != nil {
		return fmt.Errorf("创建域名池条目: %w", err)
	}
	return nil
}

// UpdateRemark 只更新备注；备注未变化时 MySQL 返回 0 行，需确认记录存在。
func (r *domainPoolRepo) UpdateRemark(ctx context.Context, id uint, remark string) error {
	res := r.db.WithContext(ctx).Model(&model.Domain{}).Where("id = ?", id).Update("remark", remark)
	if res.Error != nil {
		return fmt.Errorf("更新域名池条目: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		var count int64
		if err := r.db.WithContext(ctx).Model(&model.Domain{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return fmt.Errorf("确认域名池条目: %w", err)
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}
	}
	return nil
}

// Delete 在同一事务里清理授权关系与域名条目。
func (r *domainPoolRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("domain_id = ?", id).Delete(&model.AdminDomain{}).Error; err != nil {
			return fmt.Errorf("删除域名授权: %w", err)
		}
		res := tx.Delete(&model.Domain{}, id)
		if res.Error != nil {
			return fmt.Errorf("删除域名池条目: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *domainPoolRepo) FindByID(ctx context.Context, id uint) (*model.Domain, error) {
	var d model.Domain
	if err := r.db.WithContext(ctx).First(&d, id).Error; err != nil {
		return nil, wrapFind("域名池条目", err)
	}
	return &d, nil
}

func (r *domainPoolRepo) FindByDomain(ctx context.Context, domain string) (*model.Domain, error) {
	var d model.Domain
	if err := r.db.WithContext(ctx).Where("domain = ?", domain).First(&d).Error; err != nil {
		return nil, wrapFind("域名池条目", err)
	}
	return &d, nil
}

func (r *domainPoolRepo) List(ctx context.Context, keyword string) ([]model.Domain, error) {
	var list []model.Domain
	q := r.db.WithContext(ctx).Model(&model.Domain{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("domain LIKE ? OR remark LIKE ?", like, like)
	}
	if err := q.Order("domain ASC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("列出域名池: %w", err)
	}
	return list, nil
}

func (r *domainPoolRepo) ListByAdmin(ctx context.Context, adminID uint) ([]model.Domain, error) {
	var list []model.Domain
	err := r.db.WithContext(ctx).Model(&model.Domain{}).
		Select("domains.*").
		Joins("JOIN admin_domains ON admin_domains.domain_id = domains.id").
		Where("admin_domains.admin_id = ?", adminID).
		Order("domains.domain ASC").
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("列出管理员可用域名: %w", err)
	}
	return list, nil
}

func (r *domainPoolRepo) ListAdminDomainIDs(ctx context.Context, adminID uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&model.AdminDomain{}).
		Where("admin_id = ?", adminID).
		Order("domain_id ASC").
		Pluck("domain_id", &ids).Error
	if err != nil {
		return nil, fmt.Errorf("列出管理员域名授权: %w", err)
	}
	return ids, nil
}

// SetAdminDomains 覆盖式写入授权：先清空再批量插入。
func (r *domainPoolRepo) SetAdminDomains(ctx context.Context, adminID uint, domainIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("admin_id = ?", adminID).Delete(&model.AdminDomain{}).Error; err != nil {
			return fmt.Errorf("清理管理员域名授权: %w", err)
		}
		if len(domainIDs) == 0 {
			return nil
		}
		rows := make([]model.AdminDomain, 0, len(domainIDs))
		for _, id := range domainIDs {
			rows = append(rows, model.AdminDomain{AdminID: adminID, DomainID: id})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return fmt.Errorf("写入管理员域名授权: %w", err)
		}
		return nil
	})
}

// ListBindings 返回域名池条目与穿透记录绑定情况的对应关系。
func (r *domainPoolRepo) ListBindings(ctx context.Context) ([]DomainBinding, error) {
	var list []DomainBinding
	// GORM 把模型里的 HTTPSEnabled 落成 http_s_enabled 列，这里必须用真实列名，
	// 再用 https_enabled 别名交给 DomainBinding 接收。
	err := r.db.WithContext(ctx).Table("domains AS d").
		Select(`d.id AS domain_id,
			COALESCE(td.tunnel_id, 0) AS tunnel_id,
			COALESCE(t.name, '') AS tunnel_name,
			COALESCE(td.http_s_enabled, 0) AS https_enabled,
			COALESCE(td.cert_status, '') AS cert_status,
			COALESCE(td.nginx_status, '') AS nginx_status`).
		Joins("LEFT JOIN tunnel_domains AS td ON td.domain = d.domain").
		Joins("LEFT JOIN tunnels AS t ON t.id = td.tunnel_id AND t.deleted_at IS NULL").
		Scan(&list).Error
	if err != nil {
		return nil, fmt.Errorf("列出域名绑定关系: %w", err)
	}
	return list, nil
}
