package config

import (
	"errors"
)

const (
	defaultSessionStore = "db"
	defaultDataProvider = "cryptocompare"
)

var errEmptyDatabaseUrl = errors.New("database url cannot be blank")

type App struct {
	RunMode      string
	DataProvider string
	SessionStore string
	ApiKey       string
	DatabaseUrl  string
	Version      string

	Http  *Http
	Redis *Redis
}

func NewAppConfig() *App {
	return &App{
		Http:  &Http{},
		Redis: &Redis{},
	}
}

func (c *App) Validate() error {
	if err := c.Http.Validate(); err != nil {
		return err
	}
	if c.SessionStore == "redis" {
		if err := c.Redis.Validate(); err != nil {
			return err
		}
	}
	if c.DatabaseUrl == "" {
		return errEmptyDatabaseUrl
	}

	return nil
}
