package storage


import (
	"context"
	"fmt"
	"time"

	redis "github.com/go-redis/redis/v8"
)

type Config struct {
	Addrs []string `mapstructure:"addresses"`
	DB    int      `mapstructure:"db"`

	Username         string `mapstructure:"username"`
	Password         string `mapstructure:"password"`
	SentinelPassword string `mapstructure:"sentinel_password"`

	MaxRetries      int           `mapstructure:"max_retries"`
	MinRetryBackoff time.Duration `mapstructure:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `mapstructure:"max_retry_backoff"`

	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`

	PoolFIFO bool `mapstructure:"pool_fifo"`

	PoolSize           int           `mapstructure:"pool_size"`
	MinIdleConns       int           `mapstructure:"min_idle_conns"`
	MaxConnAge         time.Duration `mapstructure:"max_conn_age"`
	PoolTimeout        time.Duration `mapstructure:"pool_timeout"`
	IdleTimeout        time.Duration `mapstructure:"idle_timeout"`
	IdleCheckFrequency time.Duration `mapstructure:"idle_check_frequency"`

	MaxRedirects   int  `mapstructure:"max_redirects"`
	ReadOnly       bool `mapstructure:"read_only"`
	RouteByLatency bool `mapstructure:"route_by_latency"`
	RouteRandomly  bool `mapstructure:"route_randomly"`

	MasterName string `mapstructure:"master_name"`
}

type Storage struct {
	Client redis.UniversalClient
}

func New(cfg Config) (*Storage, error) {
	rdsClient := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: cfg.Addrs,
		DB:    cfg.DB,

		Username:         cfg.Username,
		Password:         cfg.Password,
		SentinelPassword: cfg.SentinelPassword,

		MaxRetries:      cfg.MaxRetries,
		MinRetryBackoff: cfg.MinRetryBackoff,
		MaxRetryBackoff: cfg.MaxRetryBackoff,

		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,

		PoolFIFO: cfg.PoolFIFO,

		PoolSize:           cfg.PoolSize,
		MinIdleConns:       cfg.MinIdleConns,
		MaxConnAge:         cfg.MaxConnAge,
		PoolTimeout:        cfg.PoolTimeout,
		IdleTimeout:        cfg.IdleTimeout,
		IdleCheckFrequency: cfg.IdleCheckFrequency,

		MaxRedirects:   cfg.MaxRedirects,
		ReadOnly:       cfg.ReadOnly,
		RouteByLatency: cfg.RouteByLatency,
		RouteRandomly:  cfg.RouteRandomly,

		MasterName: cfg.MasterName,
	})

	if err := rdsClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	s := &Storage{
		Client: rdsClient,
	}
	return s, nil
}

func (s *Storage) Close() error {
	return s.Client.Close()
}
