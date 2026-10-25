package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
)

// AvailableDomain 当前管理员可绑定、且尚未被占用的域名。
type AvailableDomain struct {
	ID     uint   `json:"id"`
	Domain string `json:"domain"`
	Remark string `json:"remark"`
}

// AvailableDomains 返回当前管理员有权使用且尚未绑定的域名。
// 超级管理员可见全部域名池，普通管理员只看被授权的域名。
func (s *DomainService) AvailableDomains(ctx context.Context, adminID uint) ([]AvailableDomain, error) {
	admin, err := s.admins.FindByID(ctx, adminID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}
	var pool []model.Domain
	if admin.IsSuperAdmin() {
		pool, err = s.pool.List(ctx, "")
	} else {
		pool, err = s.pool.ListByAdmin(ctx, adminID)
	}
	if err != nil {
		return nil, err
	}
	bound, err := s.domains.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	used := make(map[string]struct{}, len(bound))
	for _, d := range bound {
		used[d.Domain] = struct{}{}
	}
	result := make([]AvailableDomain, 0, len(pool))
	for _, d := range pool {
		if _, taken := used[d.Domain]; taken {
			continue
		}
		result = append(result, AvailableDomain{ID: d.ID, Domain: d.Domain, Remark: d.Remark})
	}
	return result, nil
}

// checkDomainAllowed 校验域名在池中且当前管理员有权使用。
func (s *DomainService) checkDomainAllowed(ctx context.Context, adminID uint, host string) error {
	entry, err := s.pool.FindByDomain(ctx, host)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDomainNotInPool
		}
		return err
	}
	admin, err := s.admins.FindByID(ctx, adminID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrForbidden
		}
		return err
	}
	if admin.IsSuperAdmin() {
		return nil
	}
	ids, err := s.pool.ListAdminDomainIDs(ctx, adminID)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if id == entry.ID {
			return nil
		}
	}
	return fmt.Errorf("%w: %s", ErrDomainNotAllowed, host)
}
