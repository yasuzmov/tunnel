package handler

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"tunnel-server/internal/pkg/response"
)

// SPA 托管内嵌的 Vue 构建产物，未知路径回退到 index.html。
func SPA(fsys fs.FS) gin.HandlerFunc {
	fileServer := http.FileServer(http.FS(fsys))
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			response.NotFound(c, "接口不存在")
			return
		}
		p := strings.TrimPrefix(pathOrIndex(c.Request.URL.Path), "/")
		if _, err := fs.Stat(fsys, p); err != nil {
			c.Request.URL.Path = "/"
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

func pathOrIndex(p string) string {
	if p == "" || p == "/" {
		return "/index.html"
	}
	return p
}
