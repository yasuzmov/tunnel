package service

import (
	"context"

	"tunnel-server/internal/model"
	"tunnel-server/internal/repository"
)

// SessionService 查询客户端联通历史。在线状态由隧道接入路径写入。
type SessionService struct {
	sessions repository.SessionRepository
}

// NewSessionService 创建联通记录查询服务。
func NewSessionService(sessions repository.SessionRepository) *SessionService {
	return &SessionService{sessions: sessions}
}

// List 按穿透记录分页返回联通历史，在线记录排在前面由仓储按 id 倒序近似满足。
func (s *SessionService) List(ctx context.Context, tunnelID uint, page, pageSize int) ([]model.ClientSession, int64, error) {
	return s.sessions.ListByTunnel(ctx, tunnelID, page, pageSize)
}
