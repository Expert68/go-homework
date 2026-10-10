package middleware

import (
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"homework04/internal/utils"

	"github.com/gin-gonic/gin"
)

// AccessLog 记录每个请求的方法、路径、状态码、耗时、IP。
// 类比 Python 的 logging 中间件 / uvicorn access log。
func AccessLog(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next() // 先让后面的逻辑跑完，再记录（所以能拿到最终的 status）
		cost := time.Since(start)

		status := c.Writer.Status()
		attrs := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency_ms", cost.Milliseconds(),
			"ip", c.ClientIP(),
		}
		if name, ok := c.Get(ContextUsername); ok {
			attrs = append(attrs, "username", name)
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, "error", c.Errors.String())
		}

		// 按严重程度分级打日志：5xx 用 Error，4xx 用 Warn，其余 Info
		switch {
		case status >= 500:
			log.Error("http_request", attrs...)
		case status >= 400:
			log.Warn("http_request", attrs...)
		default:
			log.Info("http_request", attrs...)
		}
	}
}

// Recovery 捕获 panic，防止一个请求崩掉整个服务。
// 类比 FastAPI 的 @app.exception_handler(Exception)。
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					"error", fmt.Sprint(r),
					"path", c.Request.URL.Path,
					"stack", string(debug.Stack()), // 打印堆栈，方便定位
				)
				utils.Fail(c, utils.Internal(fmt.Errorf("panic: %v", r)))
				c.Abort()
			}
		}()
		c.Next()
	}
}
