package config

import (
	"os"
	"time"
)

type Config struct {
	Port      string
	DBPath    string
	JWTSecret []byte
	JWTTTL    time.Duration
	Debug     bool
	SeedDemo  bool
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		DBPath:    getEnv("DB_PATH", "blog.db"),
		JWTSecret: []byte(getEnv("JWT_SECRET", "dev-secret")),
		JWTTTL:    time.Hour * 12,
		Debug:     getEnv("DEBUG", "") == "1",
		SeedDemo:  getEnv("SEED_DEMO", "") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
