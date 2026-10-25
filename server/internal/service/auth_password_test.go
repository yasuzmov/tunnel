package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
	"tunnel-server/internal/pkg/crypto"
)

type fakeAdminRepo struct {
	user *model.AdminUser
}

func (f *fakeAdminRepo) FindByUsername(_ context.Context, name string) (*model.AdminUser, error) {
	if f.user == nil || f.user.Username != name {
		return nil, gorm.ErrRecordNotFound
	}
	return f.user, nil
}

func (f *fakeAdminRepo) FindByID(_ context.Context, id uint) (*model.AdminUser, error) {
	if f.user == nil || f.user.ID != id {
		return nil, gorm.ErrRecordNotFound
	}
	return f.user, nil
}

func (f *fakeAdminRepo) UpdatePassword(_ context.Context, id uint, hash string) error {
	if f.user == nil || f.user.ID != id {
		return gorm.ErrRecordNotFound
	}
	f.user.PasswordHash = hash
	return nil
}

func (f *fakeAdminRepo) Create(_ context.Context, u *model.AdminUser) error {
	f.user = u
	return nil
}

func (f *fakeAdminRepo) UpdateProfile(_ context.Context, id uint, username, role string) error {
	if f.user == nil || f.user.ID != id {
		return gorm.ErrRecordNotFound
	}
	f.user.Username = username
	f.user.Role = role
	return nil
}

func (f *fakeAdminRepo) Delete(_ context.Context, id uint) error {
	if f.user == nil || f.user.ID != id {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (f *fakeAdminRepo) List(_ context.Context) ([]model.AdminUser, error) {
	if f.user == nil {
		return nil, nil
	}
	return []model.AdminUser{*f.user}, nil
}

func (f *fakeAdminRepo) CountByRole(_ context.Context, role string) (int64, error) {
	if f.user != nil && f.user.Role == role {
		return 1, nil
	}
	return 0, nil
}

func (f *fakeAdminRepo) ListMenus(_ context.Context, _ uint) ([]string, error) {
	return nil, nil
}

func (f *fakeAdminRepo) SetMenus(_ context.Context, _ uint, _ []string) error {
	return nil
}

func newAuthWithUser(t *testing.T, password string) (*AuthService, *fakeAdminRepo) {
	t.Helper()
	hash, err := crypto.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	repo := &fakeAdminRepo{user: &model.AdminUser{ID: 1, Username: "admin", PasswordHash: hash}}
	return NewAuthService(repo, "secret", time.Hour), repo
}

func TestChangePasswordSuccess(t *testing.T) {
	svc, repo := newAuthWithUser(t, "admin123")
	err := svc.ChangePassword(context.Background(), 1, ChangePasswordInput{OldPassword: "admin123", NewPassword: "newpass99"})
	if err != nil {
		t.Fatal(err)
	}
	if !crypto.CheckPassword(repo.user.PasswordHash, "newpass99") {
		t.Fatal("new password not stored")
	}
	if _, err := svc.Login(context.Background(), "admin", "newpass99"); err != nil {
		t.Fatalf("login with new password: %v", err)
	}
}

func TestChangePasswordWrongOld(t *testing.T) {
	svc, _ := newAuthWithUser(t, "admin123")
	err := svc.ChangePassword(context.Background(), 1, ChangePasswordInput{OldPassword: "bad", NewPassword: "newpass99"})
	if !errors.Is(err, ErrWrongOldPassword) {
		t.Fatalf("want ErrWrongOldPassword, got %v", err)
	}
}

func TestChangePasswordRejectsWeak(t *testing.T) {
	svc, _ := newAuthWithUser(t, "admin123")
	cases := []ChangePasswordInput{
		{OldPassword: "admin123", NewPassword: "short"},
		{OldPassword: "admin123", NewPassword: "admin123"},
		{OldPassword: "", NewPassword: "newpass99"},
	}
	for _, in := range cases {
		if err := svc.ChangePassword(context.Background(), 1, in); !errors.Is(err, ErrBadInput) {
			t.Fatalf("input %+v: want ErrBadInput, got %v", in, err)
		}
	}
}
