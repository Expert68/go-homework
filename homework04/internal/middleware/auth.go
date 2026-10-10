package middleware

import (
	"homework04/internal/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID   = "current_user_id"
	ContextUsername = "current_username"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			utils.Fail(c, utils.Unauthorized("缺少Authorization，请先登录"))
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			utils.Fail(c, utils.Unauthorized("Authorization 格式应为: Bearer <token>"))
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			utils.Fail(c, utils.Unauthorized("令牌无效或已过期"))
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUsername, claims.Username)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) uint {
	return c.GetUint(ContextUserID)
}
