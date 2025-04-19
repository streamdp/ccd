package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/go-redis/redis"
)

var (
	errConfigNotInitialized = errors.New("config not initialized")
	errRedisHost            = errors.New("redis host couldn't be blank")
	errRedisDb              = errors.New("redis db variable should be in interval 0..15")
)

type Redis struct {
	Host     string
	Port     int
	Password string
	Db       int
}

func (r *Redis) Options() (*redis.Options, error) {
	if redisUrl := os.Getenv("REDIS_URL"); redisUrl != "" {
		options, err := redis.ParseURL(redisUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to parse redis url: %w", err)
		}

		return options, nil
	}

	if r == nil {
		return nil, errConfigNotInitialized
	}

	if h := os.Getenv("REDIS_HOSTNAME"); h != "" {
		r.Host = h
	}

	if p := os.Getenv("REDIS_PORT"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_PORT: %w", errWrongNetworkPort)
		}
		r.Port = n
	}

	if pass := os.Getenv("REDIS_PASSWORD"); pass != "" {
		r.Password = pass
	}

	if d := os.Getenv("REDIS_DB"); d != "" {
		n, err := strconv.Atoi(d)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_DB: %w", errRedisDb)
		}
		r.Db = n
	}

	return &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", r.Host, r.Port),
		Password: r.Password,
		DB:       r.Db,
	}, nil
}

func (r *Redis) Validate() error {
	if r.Host == "" {
		return errRedisHost
	}
	if r.Port < 0 || r.Port > 65535 {
		return fmt.Errorf("redis: %w", errWrongNetworkPort)
	}
	if r.Db < 0 || r.Db > 15 {
		return fmt.Errorf("redis: %w", errRedisDb)
	}

	return nil
}
