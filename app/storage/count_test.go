package storage

import (
	"context"
	"sync"
	"time"
)

func (s *TestSuite) TestIncr() {
	ctx := context.Background()

	cnt := int64(2000)
	size := int64(2)
	total := cnt * size
	ttl := time.Hour

	var wg sync.WaitGroup
	for i := int64(0); i < cnt; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.Storage.Incr(ctx, "counter:12345", "1", size, ttl)
			s.Assert().NoError(err)

			time.Sleep(time.Second)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.Storage.Incr(ctx, "counter:12345", "0", size, ttl)
			s.Assert().NoError(err)

			time.Sleep(time.Second)
		}()
	}
	wg.Wait()

	for _, tt := range []struct {
		key, field string
		val        int64
	}{
		{"counter:12345", "0", total},
		{"counter:12345", "1", total},
	} {
		n, err := s.Storage.Client.HGet(ctx, tt.key, tt.field).Int64()
		s.Assert().NoError(err)
		s.Assert().Equal(tt.val, n)
	}
}
