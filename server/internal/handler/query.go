// Package handler 暴露管理 API。只做参数绑定、鉴权上下文与响应封装。
package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func queryInt(c *gin.Context, key string, def int) int {
	v := c.Query(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func queryUint(c *gin.Context, key string) (uint, bool) {
	v := c.Param(key)
	if v == "" {
		v = c.Query(key)
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(n), true
}

func queryTime(c *gin.Context, key string) *time.Time {
	v := c.Query(key)
	if v == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil
	}
	return &t
}

func optionalStatus(c *gin.Context) *int {
	v := c.Query("status")
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	return &n
}
