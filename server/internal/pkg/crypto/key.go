// Package crypto 提供接入 key 与口令哈希，避免业务层直接依赖底层随机数实现。
package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// NewTunnelKey 生成 32 字节随机 key，以 hex 表示（64 字符）。
// 足够长以防止枚举，同时方便复制粘贴到客户端。
func NewTunnelKey() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("生成隧道 key: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
