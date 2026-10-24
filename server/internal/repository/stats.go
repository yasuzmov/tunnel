package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
)

type statsRepo struct {
	db *gorm.DB
}

// NewStatsRepo 创建统计仓储。
func NewStatsRepo(db *gorm.DB) StatsRepository {
	return &statsRepo{db: db}
}

func (r *statsRepo) Dashboard(ctx context.Context, todayStart time.Time) (DashboardStats, error) {
	var s DashboardStats
	db := r.db.WithContext(ctx)
	if err := db.Model(&model.ClientSession{}).Where("online = ?", true).Count(&s.OnlineSessions).Error; err != nil {
		return s, fmt.Errorf("统计在线数: %w", err)
	}
	if err := db.Model(&model.Tunnel{}).Count(&s.TunnelCount).Error; err != nil {
		return s, fmt.Errorf("统计穿透记录: %w", err)
	}
	if err := db.Model(&model.RequestLog{}).Where("created_at >= ?", todayStart).Count(&s.TodayRequests).Error; err != nil {
		return s, fmt.Errorf("统计今日请求: %w", err)
	}
	errQ := db.Model(&model.RequestLog{}).Where("created_at >= ? AND (status >= ? OR error_message <> '')", todayStart, 400)
	if err := errQ.Count(&s.TodayErrors).Error; err != nil {
		return s, fmt.Errorf("统计今日错误: %w", err)
	}
	if s.TodayRequests > 0 {
		s.ErrorRate = float64(s.TodayErrors) / float64(s.TodayRequests)
	}
	return s, nil
}
