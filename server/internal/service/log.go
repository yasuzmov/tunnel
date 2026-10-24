package service

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
	"tunnel-server/internal/repository"
)

// LogService 请求日志查询与过期清理。
type LogService struct {
	logs       repository.LogRepository
	retainDays int
}

// NewLogService 创建日志服务。
func NewLogService(logs repository.LogRepository, retainDays int) *LogService {
	if retainDays <= 0 {
		retainDays = 30
	}
	return &LogService{logs: logs, retainDays: retainDays}
}

// List 按筛选条件分页。
func (s *LogService) List(ctx context.Context, f repository.LogFilter) ([]model.RequestLog, int64, error) {
	return s.logs.List(ctx, f)
}

// Get 读取含请求/响应体的详情。
func (s *LogService) Get(ctx context.Context, id uint) (*model.RequestLog, error) {
	item, err := s.logs.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return item, nil
}

// CleanupExpired 删除超过保留天数的日志。请求体表通过外键或应用层级联，这里只删主表。
func (s *LogService) CleanupExpired(ctx context.Context) (int64, error) {
	before := time.Now().AddDate(0, 0, -s.retainDays)
	return s.logs.DeleteOlderThan(ctx, before)
}
