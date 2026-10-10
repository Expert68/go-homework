package database

import (
	"fmt"
	"homework04/internal/config"
	"homework04/internal/models"
	"log/slog"
	"time"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.Config, log *slog.Logger) error {
	gormLogLevel := logger.Warn
	if cfg.Debug {
		gormLogLevel = logger.Info
	}

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{
		Logger:         logger.Default.LogMode(gormLogLevel),
		TranslateError: true,
	})

	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接池失败: %w", err)
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库ping失败: %w", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{}); err != nil {
		return fmt.Errorf("自动创建表格失败: %w", err)
	}

	DB = db

	if cfg.SeedDemo {
		if err := seed(log); err != nil {
			return fmt.Errorf("写入演示数据失败: %w", err)
		}
	}
	return nil
}

func seed(log *slog.Logger) error {
	var count int64
	if err := DB.Model(&models.User{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil //user表中已有数据
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("demo123456"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := models.User{
		Username: "demo",
		Password: string(hash),
		Email:    "demo@x.com",
	}
	if err := DB.Create(&user).Error; err != nil {
		return err
	}

	posts := []models.Post{
		{Title: "欢迎来到我的博客", Content: "这是系统自动生成的演示数据。", UserID: user.ID},
		{Title: "Gin + GORM 学习笔记", Content: "Gin 负责 HTTP，GORM 负责数据库。", UserID: user.ID},
	}
	if err := DB.Create(&posts).Error; err != nil {
		return err
	}

	log.Info("已写入演示数据", "username", "demo", "password", "demo123456")
	return nil

}
