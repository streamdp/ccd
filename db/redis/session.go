package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/streamdp/ccd/config"
)

const sessionName = "lastSession"

type keysStore struct {
	c *redis.Client
}

var errKeyStoreNotInitialized = errors.New("key store not initialised")

// NewRedisKeysStore initialize new redis session store
func NewRedisKeysStore(cfg *config.App) (*keysStore, error) {
	opt, err := cfg.Redis.Options()
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis os environment variables: %w", err)
	}

	client := redis.NewClient(opt)
	if _, err = client.Ping().Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &keysStore{
		c: client,
	}, nil
}

// GetSession get previously saved session
func (s *keysStore) GetSession(ctx context.Context) (map[string]int64, error) {
	if s == nil {
		return nil, errKeyStoreNotInitialized
	}
	session := make(map[string]int64)
	for k, v := range s.c.WithContext(ctx).HGetAll(sessionName).Val() {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to get session: %w", err)
		}
		session[k] = i
	}

	return session, nil
}

// AddTask add a new task or update an already saved task in the current session
func (s *keysStore) AddTask(ctx context.Context, n string, i int64) error {
	if s == nil {
		return errKeyStoreNotInitialized
	}

	s.c.WithContext(ctx).HSet(sessionName, n, strconv.FormatInt(i, 10))

	return nil
}

// UpdateTask add a new task or update an already saved task in the current session
func (s *keysStore) UpdateTask(ctx context.Context, n string, i int64) error {
	return s.AddTask(ctx, n, i)
}

// RemoveTask remove a task from the current session
func (s *keysStore) RemoveTask(ctx context.Context, n string) error {
	if s == nil {
		return errKeyStoreNotInitialized
	}

	s.c.WithContext(ctx).HDel(sessionName, n)

	return nil
}

func (s *keysStore) Close() error {
	if s.c == nil {
		return errKeyStoreNotInitialized
	}

	if err := s.c.Close(); err != nil {
		return fmt.Errorf("failed to close key store: %w", err)
	}

	return nil
}
