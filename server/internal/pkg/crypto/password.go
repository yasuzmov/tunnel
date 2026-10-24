package crypto

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 10

// HashPassword 使用 bcrypt 哈希管理员口令。
func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("哈希口令: %w", err)
	}
	return string(hashed), nil
}

// CheckPassword 比对明文与哈希。错误统一视为不匹配，避免把 bcrypt 细节泄漏给调用方。
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
