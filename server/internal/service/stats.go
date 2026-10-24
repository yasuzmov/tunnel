package service

import (
	"context"
	"time"

	"tunnel-server/internal/repository"
)

// StatsService 仪表盘聚合。
type StatsService struct {
	stats repository.StatsRepository
}

// NewStatsService 创建统计服务。
func NewStatsService(stats repository.StatsRepository) *StatsService {
	return &StatsService{stats: stats}
}

// Dashboard 返回今日零点起的汇总。
func (s *StatsService) Dashboard(ctx context.Context) (repository.DashboardStats, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return s.stats.Dashboard(ctx, today)
}
