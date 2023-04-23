package storage

import (
	"context"
	"fmt"
	"time"

	redis "github.com/go-redis/redis/v8"
)

type Config struct {
	Addrs    []string `mapstructure:"addresses"`
	DB       int      `mapstructure:"db"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`

	KeyExpireIn time.Duration `mapstructure:"key_expire_in"`
}

type Storage struct {
	Expire time.Duration
	Client redis.UniversalClient
}

func New(cfg Config) (*Storage, error) {
	rdsClient := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    cfg.Addrs,
		DB:       cfg.DB,
		Username: cfg.Username,
		Password: cfg.Password,
	})

	if err := rdsClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	expire := 24 * time.Hour
	if cfg.KeyExpireIn > 0 {
		expire = cfg.KeyExpireIn
	}
	s := &Storage{
		Expire: expire,
		Client: rdsClient,
	}
	return s, nil
}

func (s *Storage) Close() error {
	return s.Client.Close()
}
