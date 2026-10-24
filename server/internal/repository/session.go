package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
)

type sessionRepo struct {
	db *gorm.DB
}

// NewSessionRepo 创建联通记录仓储。
func NewSessionRepo(db *gorm.DB) SessionRepository {
	return &sessionRepo{db: db}
}

func (r *sessionRepo) Create(ctx context.Context, s *model.ClientSession) error {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return fmt.Errorf("创建联通记录: %w", err)
	}
	return nil
}

func (r *sessionRepo) Update(ctx context.Context, s *model.ClientSession) error {
	if err := r.db.WithContext(ctx).Save(s).Error; err != nil {
		return fmt.Errorf("更新联通记录: %w", err)
	}
	return nil
}

func (r *sessionRepo) FindByID(ctx context.Context, id uint) (*model.ClientSession, error) {
	var s model.ClientSession
	if err := r.db.WithContext(ctx).First(&s, id).Error; err != nil {
		return nil, wrapFind("联通记录", err)
	}
	return &s, nil
}

func (r *sessionRepo) ListByTunnel(ctx context.Context, tunnelID uint, page, pageSize int) ([]model.ClientSession, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	q := r.db.WithContext(ctx).Model(&model.ClientSession{}).Where("tunnel_id = ?", tunnelID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计联通记录: %w", err)
	}
	var list []model.ClientSession
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, fmt.Errorf("列出联通记录: %w", err)
	}
	return list, total, nil
}

func (r *sessionRepo) TouchHeartbeat(ctx context.Context, id uint, at time.Time) error {
	err := r.db.WithContext(ctx).Model(&model.ClientSession{}).Where("id = ?", id).
		Update("last_heartbeat_at", at).Error
	if err != nil {
		return fmt.Errorf("更新联通心跳: %w", err)
	}
	return nil
}

func (r *sessionRepo) MarkOffline(ctx context.Context, id uint, at time.Time, reason string, stats SessionStats) error {
	err := r.db.WithContext(ctx).Model(&model.ClientSession{}).Where("id = ?", id).Updates(map[string]any{
		"online":            false,
		"disconnected_at":   at,
		"disconnect_reason": reason,
		"request_count":     stats.RequestCount,
		"bytes_in":          stats.BytesIn,
		"bytes_out":         stats.BytesOut,
	}).Error
	if err != nil {
		return fmt.Errorf("标记联通离线: %w", err)
	}
	return nil
}

func (r *sessionRepo) CountOnline(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.ClientSession{}).Where("online = ?", true).Count(&n).Error
	if err != nil {
		return 0, fmt.Errorf("统计在线会话: %w", err)
	}
	return n, nil
}
