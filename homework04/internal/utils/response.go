package utils

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(200, Response{
		Code:    int(CodeOK),
		Message: "OK",
		Data:    data,
	})
}

func Created(c *gin.Context, data any) {
	c.JSON(200, Response{
		Code:    int(CodeOK),
		Message: "Created",
		Data:    data,
	})
}

func Fail(c *gin.Context, err error) {
	appErr := AsAppError(err)
	_ = c.Error(err)
	c.JSON(appErr.Code.HTTPStatus(),
		Response{
			Code:    int(appErr.Code),
			Message: appErr.Message,
		})
}
