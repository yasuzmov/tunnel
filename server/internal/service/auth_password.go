package service

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"gorm.io/gorm"

	"tunnel-server/internal/pkg/crypto"
)

const (
	minPasswordLen = 8
	maxPasswordLen = 64
)

// ChangePasswordInput 修改口令参数。
type ChangePasswordInput struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ChangePassword 校验原口令后写入新口令哈希。
// 新口令长度 8~64 且不能与原口令相同；bcrypt 只取前 72 字节，上限留出余量。
// 已签发的 JWT 不会失效，前端在成功后主动退出登录。
func (s *AuthService) ChangePassword(ctx context.Context, userID uint, in ChangePasswordInput) error {
	if err := validateNewPassword(in.OldPassword, in.NewPassword); err != nil {
		return err
	}
	user, err := s.admins.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("读取当前用户: %w", err)
	}
	if !crypto.CheckPassword(user.PasswordHash, in.OldPassword) {
		return ErrWrongOldPassword
	}
	hashed, err := crypto.HashPassword(in.NewPassword)
	if err != nil {
		return err
	}
	if err := s.admins.UpdatePassword(ctx, userID, hashed); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func validateNewPassword(oldPwd, newPwd string) error {
	if oldPwd == "" || newPwd == "" {
		return fmt.Errorf("%w: 请填写原密码和新密码", ErrBadInput)
	}
	if err := validatePassword(newPwd); err != nil {
		return err
	}
	if oldPwd == newPwd {
		return fmt.Errorf("%w: 新密码不能与原密码相同", ErrBadInput)
	}
	return nil
}

// validatePassword 校验口令长度；bcrypt 只取前 72 字节，上限留出余量。
func validatePassword(pwd string) error {
	n := utf8.RuneCountInString(pwd)
	if n < minPasswordLen || n > maxPasswordLen {
		return fmt.Errorf("%w: 密码长度需为 %d~%d 位", ErrBadInput, minPasswordLen, maxPasswordLen)
	}
	return nil
}
