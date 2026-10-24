package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"tunnel-server/internal/pkg/jwtutil"
	"tunnel-server/internal/pkg/response"
)

const ctxClaims = "auth.claims"

// JWT 校验 Authorization Bearer 或 query token（供 EventSource 使用）。
func JWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c)
		if token == "" {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		claims, err := jwtutil.Parse(secret, token)
		if err != nil {
			response.Unauthorized(c, "登录已失效")
			c.Abort()
			return
		}
		c.Set(ctxClaims, claims)
		c.Next()
	}
}

// Claims 取出当前登录身份。
func Claims(c *gin.Context) *jwtutil.Claims {
	v, ok := c.Get(ctxClaims)
	if !ok {
		return nil
	}
	claims, _ := v.(*jwtutil.Claims)
	return claims
}

func bearerToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return strings.TrimSpace(c.Query("token"))
}
