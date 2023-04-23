package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestSuite struct {
	suite.Suite
	redis *redisContainer
	*Storage
}

func (s *TestSuite) SetupSuite() {
	cfg := Config{
		Addrs:       []string{s.redis.Addr},
		KeyExpireIn: time.Hour,
	}
	storage, err := New(cfg)
	s.NoError(err)
	s.Storage = storage
}

func (s *TestSuite) TearDownSuite() {
	s.NoError(s.Storage.Close())
}

func (s *TestSuite) SetupTest() {
	ctx := context.Background()
	err := s.Storage.Client.FlushAll(ctx).Err()
	s.NoError(err)
}

func TestRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	redisContainer, err := startRedisContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	suite.Run(t, &TestSuite{
		redis: redisContainer,
	})
}

type redisContainer struct {
	testcontainers.Container
	Addr string
}

func startRedisContainer(ctx context.Context) (*redisContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:5",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("* Ready to accept connections"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())
	return &redisContainer{Container: container, Addr: addr}, nil
}
