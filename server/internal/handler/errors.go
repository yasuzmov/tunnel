package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"tunnel-server/internal/pkg/response"
	"tunnel-server/internal/service"
)

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrBadInput):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrTunnelOffline):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrDomainTaken):
		response.Fail(c, 409, 409, err.Error())
	case errors.Is(err, service.ErrDomainMismatch):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrInvalidCredential):
		response.Unauthorized(c, err.Error())
	case errors.Is(err, service.ErrOperation):
		response.Fail(c, 500, 500, err.Error())
	default:
		response.ServerError(c, "操作失败")
	}
}
