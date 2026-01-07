package config

import (
	"os"
	"time"
)

type RedisConfig struct {
	Addr        string        `env:"REDIS_ADDR"`
	Password    string        `env:"REDIS_PASSWORD"`
	User        string        `env:"REDIS_USER"`
	DB          int           `env:"REDIS_DB"`
	MaxRetries  int           `env:"REDIS_MAX_RETRIES"`
	DialTimeout time.Duration `env:"REDIS_DIAL_TIMEOUT"`
	Timeout     time.Duration `env:"REDIS_TIMEOUT"`
}

func GetDBURL() string {
	dbURL := os.Getenv("DATABASE_URL")
	return dbURL
}

func GetRedisConfig() RedisConfig {
	redisAddr := os.Getenv("REDIS_ADDR")
	redisConfig := RedisConfig{
		Addr: redisAddr,
	}
	return redisConfig
}
