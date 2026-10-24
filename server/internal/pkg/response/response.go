// Package response 统一管理 API 的 JSON 外壳。
// Handler 只调用本包，避免各处拼不同的 code/message 字段。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope 是所有管理 API 的响应结构。code=0 表示成功。
type Envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// PageData 分页列表。
type PageData struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// OK 返回成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Code: 0, Message: "ok", Data: data})
}

// Page 返回分页成功响应。
func Page(c *gin.Context, list any, total int64, page, pageSize int) {
	OK(c, PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

// Fail 返回业务或校验失败。httpStatus 与业务 code 分开，便于前端统一处理 401。
func Fail(c *gin.Context, httpStatus, code int, message string) {
	c.JSON(httpStatus, Envelope{Code: code, Message: message})
}

// BadRequest 表示参数不合法。
func BadRequest(c *gin.Context, message string) {
	Fail(c, http.StatusBadRequest, 400, message)
}

// Unauthorized 表示未登录或 token 失效。
func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, 401, message)
}

// Forbidden 表示无权限。
func Forbidden(c *gin.Context, message string) {
	Fail(c, http.StatusForbidden, 403, message)
}

// NotFound 表示资源不存在。
func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, 404, message)
}

// ServerError 对外隐藏内部细节，只给一句可展示文案。
func ServerError(c *gin.Context, message string) {
	if message == "" {
		message = "服务内部错误"
	}
	Fail(c, http.StatusInternalServerError, 500, message)
}
