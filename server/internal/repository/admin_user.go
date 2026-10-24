package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
)

type adminUserRepo struct {
	db *gorm.DB
}

// NewAdminUserRepo 创建管理员仓储。
func NewAdminUserRepo(db *gorm.DB) AdminUserRepository {
	return &adminUserRepo{db: db}
}

// FindByUsername 按登录名查找，未找到返回 gorm.ErrRecordNotFound。
func (r *adminUserRepo) FindByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	var u model.AdminUser
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, wrapFind("管理员", err)
	}
	return &u, nil
}

// FindByID 按主键查找。
func (r *adminUserRepo) FindByID(ctx context.Context, id uint) (*model.AdminUser, error) {
	var u model.AdminUser
	err := r.db.WithContext(ctx).First(&u, id).Error
	if err != nil {
		return nil, wrapFind("管理员", err)
	}
	return &u, nil
}

func wrapFind(name string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return fmt.Errorf("查询%s: %w", name, err)
}
