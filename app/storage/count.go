package storage

import (
	"context"
	"time"

	redis "github.com/go-redis/redis/v8"
)

func (s *Storage) Incr(ctx context.Context, key string, field string, val int64, exp time.Duration) error {
	_, err := s.Client.Pipelined(ctx, func(pip redis.Pipeliner) error {
		pip.HIncrBy(ctx, key, field, val)
		pip.Expire(ctx, key, exp)
		return nil
	})

	return err
}
