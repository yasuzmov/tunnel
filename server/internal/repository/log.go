package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
)

type logRepo struct {
	db *gorm.DB
}

// NewLogRepo 创建请求日志仓储。
func NewLogRepo(db *gorm.DB) LogRepository {
	return &logRepo{db: db}
}

func (r *logRepo) CreateBatch(ctx context.Context, logs []*model.RequestLog, bodies []*model.RequestLogBody) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&logs).Error; err != nil {
			return fmt.Errorf("批量写入请求日志: %w", err)
		}
		if len(bodies) == 0 {
			return nil
		}
		for i := range bodies {
			if i < len(logs) {
				bodies[i].RequestLogID = logs[i].ID
			}
		}
		if err := tx.Create(&bodies).Error; err != nil {
			return fmt.Errorf("批量写入请求体: %w", err)
		}
		return nil
	})
}

func (r *logRepo) List(ctx context.Context, f LogFilter) ([]model.RequestLog, int64, error) {
	page, pageSize := normalizePage(f.Page, f.PageSize)
	q := applyLogFilter(r.db.WithContext(ctx).Model(&model.RequestLog{}), f)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计请求日志: %w", err)
	}
	var list []model.RequestLog
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, fmt.Errorf("列出请求日志: %w", err)
	}
	return list, total, nil
}

func (r *logRepo) GetByID(ctx context.Context, id uint) (*model.RequestLog, error) {
	var item model.RequestLog
	err := r.db.WithContext(ctx).Preload("Body").First(&item, id).Error
	if err != nil {
		return nil, wrapFind("请求日志", err)
	}
	return &item, nil
}

func (r *logRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	db := r.db.WithContext(ctx)
	sub := db.Model(&model.RequestLog{}).Select("id").Where("created_at < ?", before)
	if err := db.Where("request_log_id IN (?)", sub).Delete(&model.RequestLogBody{}).Error; err != nil {
		return 0, fmt.Errorf("清理过期请求体: %w", err)
	}
	res := db.Where("created_at < ?", before).Delete(&model.RequestLog{})
	if res.Error != nil {
		return 0, fmt.Errorf("清理过期日志: %w", res.Error)
	}
	return res.RowsAffected, nil
}

func applyLogFilter(q *gorm.DB, f LogFilter) *gorm.DB {
	if f.TunnelID != nil {
		q = q.Where("tunnel_id = ?", *f.TunnelID)
	}
	if f.Domain != "" {
		q = q.Where("domain = ?", f.Domain)
	}
	if f.Method != "" {
		q = q.Where("method = ?", f.Method)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	if f.Path != "" {
		q = q.Where("path LIKE ?", "%"+f.Path+"%")
	}
	if f.RemoteIP != "" {
		q = q.Where("remote_ip = ?", f.RemoteIP)
	}
	if f.RequestID != "" {
		q = q.Where("request_id = ?", f.RequestID)
	}
	if f.Start != nil {
		q = q.Where("created_at >= ?", *f.Start)
	}
	if f.End != nil {
		q = q.Where("created_at <= ?", *f.End)
	}
	return q
}
