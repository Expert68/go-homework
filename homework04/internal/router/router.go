package router

import (
	"homework04/internal/handlers"
	"homework04/internal/middleware"
	"homework04/internal/utils"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func New(log *slog.Logger) *gin.Engine {
	r := gin.New()
	r.SetTrustedProxies(nil)

	r.Use(middleware.Recovery(log))
	r.Use(middleware.AccessLog(log))

	api := r.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
		auth := r.Group("/auth")
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
		api.GET("/posts", handlers.ListPosts)
		api.GET("/posts/:id", handlers.GetPost)
		api.GET("/posts/:id/comments", handlers.ListComments)

		authed := api.Group("", middleware.JWTAuth())
		{
			authed.GET("/auth/me", handlers.Me)
			authed.POST("/posts", handlers.CreatePost)
			authed.PUT("/posts/:id", handlers.UpdatePost)
			authed.DELETE("/posts/:id", handlers.DeletePost)
			authed.POST("/posts/:id/comments", handlers.CreateComment)
		}
	}
	r.NoRoute(func(c *gin.Context) {
		utils.Fail(c, utils.NotFound("接口地址不存在: "+c.Request.URL.Path))
	})
	return r
}
