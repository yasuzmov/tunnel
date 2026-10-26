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
// 日志按所属穿透记录的创建者隔离：普通管理员只能查看自己创建的记录产生的日志。
type LogService struct {
	logs       repository.LogRepository
	tunnels    repository.TunnelRepository
	admins     repository.AdminUserRepository
	retainDays int
}

// NewLogService 创建日志服务。tunnels 与 admins 用于按管理员隔离日志。
func NewLogService(logs repository.LogRepository, tunnels repository.TunnelRepository, admins repository.AdminUserRepository, retainDays int) *LogService {
	if retainDays <= 0 {
		retainDays = 30
	}
	return &LogService{logs: logs, tunnels: tunnels, admins: admins, retainDays: retainDays}
}

// List 按筛选条件分页；普通管理员强制附加创建者过滤，调用方传入的 CreatedBy 会被覆盖。
func (s *LogService) List(ctx context.Context, adminID uint, f repository.LogFilter) ([]model.RequestLog, int64, error) {
	v, err := resolveViewer(ctx, s.admins, adminID)
	if err != nil {
		return nil, 0, err
	}
	f.CreatedBy = v.ownerFilter()
	return s.logs.List(ctx, f)
}

// Get 读取含请求/响应体的详情；普通管理员读取他人记录的日志时返回 ErrNotFound。
func (s *LogService) Get(ctx context.Context, adminID, id uint) (*model.RequestLog, error) {
	v, err := resolveViewer(ctx, s.admins, adminID)
	if err != nil {
		return nil, err
	}
	item, err := s.logs.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if !v.Super {
		owner, err := s.tunnels.OwnerOf(ctx, item.TunnelID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		if !v.Owns(owner) {
			return nil, ErrNotFound
		}
	}
	return item, nil
}

// CleanupExpired 删除超过保留天数的日志。请求体表通过外键或应用层级联，这里只删主表。
func (s *LogService) CleanupExpired(ctx context.Context) (int64, error) {
	before := time.Now().AddDate(0, 0, -s.retainDays)
	return s.logs.DeleteOlderThan(ctx, before)
}
