// Package idgen 生成请求级唯一 ID。
package idgen

import "github.com/google/uuid"

// NewRequestID 返回 UUID v4 字符串，用于日志关联与排障。
func NewRequestID() string {
	return uuid.NewString()
}
