package main

import (
	"context"
	"errors"
	"homework04/internal/config"
	"homework04/internal/database"
	"homework04/internal/logging"
	"homework04/internal/router"
	"homework04/internal/utils"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()

	log := logging.New(cfg)

	utils.InitJWT(cfg.JWTSecret, cfg.JWTTTL)

	if err := database.Init(cfg, log); err != nil {
		log.Error("数据库初始化失败", "error", err)
		os.Exit(1)
	}
	log.Info("数据库连接成功", "Path", cfg.DBPath)
	engine := router.New(log)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("HTTP 服务已启动", "url", "http://localhost:"+cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP 服务启动失败", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("收到退出信号，开始优雅关闭...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("优雅关闭失败", "error", err)
	}
	log.Info("服务已退出")
}
