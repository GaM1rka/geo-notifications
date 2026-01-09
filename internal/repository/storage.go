package repository

import (
	"context"
	"database/sql"
	"fmt"
	"geo-notifications/internal/config"
	"geo-notifications/internal/model"
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
		Password:     cfg.Password,
		Username:     cfg.User,
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

func (s *Storage) CreateTables(ctx context.Context) error {
	query := `
CREATE TABLE IF NOT EXISTS incidents (
    id          SERIAL PRIMARY KEY,
    title       TEXT        NOT NULL,
    description TEXT        NOT NULL,
    latitude    DOUBLE PRECISION NOT NULL,
    longitude   DOUBLE PRECISION NOT NULL,
    radius_m    INTEGER     NOT NULL,
    active      BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`
	_, err := s.repo.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("create table incidents: %w", err)
	}
	return nil
}

func (s *Storage) Create(ctx context.Context, in *model.Incident) (int64, error) {
	query := `
INSERT INTO incidents (title, description, latitude, longitude, radius_m, active)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at, updated_at;
`

	row := s.repo.db.QueryRowContext(ctx, query,
		in.Title,
		in.Description,
		in.Latitude,
		in.Longitude,
		in.RadiusM,
		in.Active,
	)

	if err := row.Scan(&in.ID, &in.CreatedAt, &in.UpdatedAt); err != nil {
		return 0, err
	}
	return in.ID, nil
}

func (s *Storage) GetList(ctx context.Context, page, pageSize int) ([]model.Incident, error) {
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`
SELECT id, title, description, latitude, longitude, radius_m, active, created_at, updated_at
FROM incidents
ORDER BY created_at DESC
LIMIT %d OFFSET %d;
`, pageSize, offset)

	rows, err := s.repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.Incident
	for rows.Next() {
		var in model.Incident
		if err := rows.Scan(
			&in.ID,
			&in.Title,
			&in.Description,
			&in.Latitude,
			&in.Longitude,
			&in.RadiusM,
			&in.Active,
			&in.CreatedAt,
			&in.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, in)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Storage) GetByID(ctx context.Context, id int64) (*model.Incident, error) {
	query := `
SELECT id, title, description, latitude, longitude, radius_m, active, created_at, updated_at
FROM incidents
WHERE id = $1;
`
	var in model.Incident
	err := s.repo.db.QueryRowContext(ctx, query, id).Scan(
		&in.ID,
		&in.Title,
		&in.Description,
		&in.Latitude,
		&in.Longitude,
		&in.RadiusM,
		&in.Active,
		&in.CreatedAt,
		&in.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &in, nil
}

func (s *Storage) Update(ctx context.Context, in *model.Incident) error {
	query := `
UPDATE incidents
SET title = $1,
    description = $2,
    latitude = $3,
    longitude = $4,
    radius_m = $5,
    active = $6,
    updated_at = NOW()
WHERE id = $7;
`
	_, err := s.repo.db.ExecContext(ctx, query,
		in.Title,
		in.Description,
		in.Latitude,
		in.Longitude,
		in.RadiusM,
		in.Active,
		in.ID,
	)
	return err
}

func (s *Storage) Deactivate(ctx context.Context, id int64) error {
	query := `
UPDATE incidents
SET active = FALSE,
    updated_at = NOW()
WHERE id = $1;
`
	_, err := s.repo.db.ExecContext(ctx, query, id)
	return err
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
