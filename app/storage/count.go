package storage

import (
	"context"

	redis "github.com/go-redis/redis/v8"
)

type Info struct {
	Key   string
	Field string
	Val   int64
}

func (s *Storage) Incr(ctx context.Context, infos ...Info) error {
	_, err := s.Client.Pipelined(ctx, func(pip redis.Pipeliner) error {
		for _, i := range infos {
			pip.HIncrBy(ctx, i.Key, i.Field, i.Val)
			pip.Expire(ctx, i.Key, s.Expire)
		}
		return nil
	})
	return err
}
