package session

import (
	"fmt"
	"log"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/streamdp/ccd/config"
)

const sessionName = "lastSession"

type KeysStore struct {
	c *redis.Client
}

type redisConfig struct {
	host     string
	port     int
	password string
	db       int
}

var c = redisConfig{
	host:     "127.0.0.1",
	port:     6379,
	password: "",
	db:       0,
}

func init() {
	if h := config.GetEnv("REDIS_HOSTNAME"); h != "" {
		c.host = h
	}
	if p := config.GetEnv("REDIS_PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			c.port = n
		}
	}
	if pass := config.GetEnv("REDIS_PASSWORD"); pass != "" {
		c.password = pass
	}
	if d := config.GetEnv("REDIS_DB"); d != "" {
		if n, err := strconv.Atoi(d); err == nil {
			c.db = n
		}
	}
}

func NewKeysStore() *KeysStore {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.host, c.port),
		Password: c.password,
		DB:       c.db,
	})
	if _, err := client.Ping().Result(); err != nil {
		log.Println(fmt.Sprintf("failed to connect redis on %s:%d", c.host, c.port))
		return nil
	}
	return &KeysStore{
		c: client,
	}
}

func (s *KeysStore) SaveSession(session map[string]string) (err error) {
	if s == nil {
		return
	}
	for k, v := range session {
		if err = s.c.HSet(sessionName, k, v).Err(); err != nil {
			return
		}
	}
	return
}

func (s *KeysStore) GetSession() map[string]string {
	if s == nil {
		return nil
	}
	return s.c.HGetAll(sessionName).Val()
}

func (s *KeysStore) AppendTaskToSession(name string, interval int64) {
	if s == nil {
		return
	}
	s.c.HSet(sessionName, name, strconv.FormatInt(interval, 10))
	return
}

func (s *KeysStore) RemoveTaskFromSession(name string) {
	if s == nil {
		return
	}
	s.c.HDel(sessionName, name)
	return
}
