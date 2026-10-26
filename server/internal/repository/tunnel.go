package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
)

type tunnelRepo struct {
	db *gorm.DB
}

// NewTunnelRepo 创建穿透记录仓储。
func NewTunnelRepo(db *gorm.DB) TunnelRepository {
	return &tunnelRepo{db: db}
}

func (r *tunnelRepo) Create(ctx context.Context, t *model.Tunnel) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return fmt.Errorf("创建穿透记录: %w", err)
	}
	return nil
}

func (r *tunnelRepo) Update(ctx context.Context, t *model.Tunnel) error {
	if err := r.db.WithContext(ctx).Save(t).Error; err != nil {
		return fmt.Errorf("更新穿透记录: %w", err)
	}
	return nil
}

func (r *tunnelRepo) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.Tunnel{}, id).Error; err != nil {
		return fmt.Errorf("删除穿透记录: %w", err)
	}
	return nil
}

func (r *tunnelRepo) FindByID(ctx context.Context, id uint) (*model.Tunnel, error) {
	var t model.Tunnel
	err := r.db.WithContext(ctx).Preload("Domains").First(&t, id).Error
	if err != nil {
		return nil, wrapFind("穿透记录", err)
	}
	return &t, nil
}

func (r *tunnelRepo) FindByKey(ctx context.Context, key string) (*model.Tunnel, error) {
	var t model.Tunnel
	err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&t).Error
	if err != nil {
		return nil, wrapFind("穿透记录", err)
	}
	return &t, nil
}

func (r *tunnelRepo) List(ctx context.Context, page, pageSize int, keyword string, createdBy *uint) ([]model.Tunnel, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	q := r.db.WithContext(ctx).Model(&model.Tunnel{})
	if createdBy != nil {
		q = q.Where("created_by = ?", *createdBy)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name LIKE ? OR remark LIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计穿透记录: %w", err)
	}
	var list []model.Tunnel
	err := q.Preload("Domains").Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, fmt.Errorf("列出穿透记录: %w", err)
	}
	return list, total, nil
}

func (r *tunnelRepo) OwnerOf(ctx context.Context, id uint) (uint, error) {
	var t model.Tunnel
	err := r.db.WithContext(ctx).Unscoped().Select("id", "created_by").First(&t, id).Error
	if err != nil {
		return 0, wrapFind("穿透记录", err)
	}
	return t.CreatedBy, nil
}

// LoggingFlags 只取 id 与开关，给入口路由缓存用。
func (r *tunnelRepo) LoggingFlags(ctx context.Context) (map[uint]bool, error) {
	var rows []model.Tunnel
	err := r.db.WithContext(ctx).Select("id", "log_enabled").Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("读取日志开关: %w", err)
	}
	flags := make(map[uint]bool, len(rows))
	for _, t := range rows {
		flags[t.ID] = t.LogEnabled
	}
	return flags, nil
}

func (r *tunnelRepo) TouchOnline(ctx context.Context, id uint, at time.Time) error {
	err := r.db.WithContext(ctx).Model(&model.Tunnel{}).Where("id = ?", id).Update("last_online_at", at).Error
	if err != nil {
		return fmt.Errorf("更新最后在线时间: %w", err)
	}
	return nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}
