package repository

import (
	"context"
	"database/sql"
	"fmt"
	"geo-notifications/internal/config"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type PostgresRepo struct {
	db *sql.DB
}
type RedisCache struct {
	cache *redis.Client
}

type Storage struct {
	repo  *PostgresRepo
	cache *RedisCache
}

func NewPostgresRepo(dbURL string) (*PostgresRepo, error) {
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresRepo{db: db}, nil
}

func NewRedisCache(ctx context.Context, cfg config.RedisConfig) (*RedisCache, error) {
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 5 * time.Second
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 3 * time.Second
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password, // может быть ""
		Username:     cfg.User,     // может быть ""
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		fmt.Printf("failed to connect to redis server: %s\n", err.Error())
		return nil, err
	}

	return &RedisCache{cache: client}, nil
}

func NewStorage(dbURL string, redisCfg config.RedisConfig) (*Storage, error) {
	postgres, err := NewPostgresRepo(dbURL)
	if err != nil {
		return nil, err
	}
	redis, err := NewRedisCache(context.Background(), redisCfg)
	if err != nil {
		return nil, err
	}
	return &Storage{
		repo:  postgres,
		cache: redis,
	}, nil
}

func (s *Storage) Close() error {
	var errPostgres, errRedis error

	if s.repo != nil && s.repo.db != nil {
		errPostgres = s.repo.db.Close()
	}
	if s.cache != nil && s.cache.cache != nil {
		errRedis = s.cache.cache.Close()
	}

	if errPostgres != nil || errRedis != nil {
		return fmt.Errorf("close errors: postgres=%v, redis=%v", errPostgres, errRedis)
	}
	return nil
}
