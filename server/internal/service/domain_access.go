package service

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// AvailableDomain 当前管理员可绑定、且尚未被占用的域名。
type AvailableDomain struct {
	ID     uint   `json:"id"`
	Domain string `json:"domain"`
	Remark string `json:"remark"`
}

// AvailableDomains 返回当前管理员可使用且尚未绑定到任何穿透记录的域名。
// 超级管理员可见全部域名池；普通管理员只能使用自己在域名管理中创建的域名。
func (s *DomainService) AvailableDomains(ctx context.Context, adminID uint) ([]AvailableDomain, error) {
	v, err := resolveViewer(ctx, s.admins, adminID)
	if err != nil {
		return nil, err
	}
	pool, err := s.pool.List(ctx, "", v.ownerFilter())
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

// checkDomainAllowed 校验域名在池中且归当前管理员所有（超级管理员不限）。
// 普通管理员使用他人或不存在的域名统一返回 ErrDomainNotAllowed，不透露域名是否属于他人。
// 是否已被其他穿透记录占用由调用方另行校验。
func (s *DomainService) checkDomainAllowed(ctx context.Context, adminID uint, host string) error {
	v, err := resolveViewer(ctx, s.admins, adminID)
	if err != nil {
		return err
	}
	entry, err := s.pool.FindByDomain(ctx, host)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if v.Super {
			return ErrDomainNotInPool
		}
		return ErrDomainNotAllowed
	}
	if !v.Owns(entry.CreatedBy) {
		return ErrDomainNotAllowed
	}
	return nil
}
