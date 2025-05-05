package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	defaultSessionStore = "db"
	defaultDataProvider = "cryptocompare"
)

var errEmptyDatabaseUrl = errors.New("database url cannot be blank")

type App struct {
	DataProvider string
	SessionStore string
	ApiKey       string
	DatabaseUrl  string

	Http  *Http
	Redis *Redis

	runMode string
	debug   bool
	version string
}

func NewAppConfig() *App {
	return &App{
		Http: &Http{},
		Redis: &Redis{
			Host:     redisDefaultHost,
			Port:     redisDefaultPort,
			Password: "",
			Db:       redisDefaultDb,
		},

		runMode: gin.ReleaseMode,
		version: version,
	}
}

func (a *App) Validate() error {
	if err := a.Http.Validate(); err != nil {
		return err
	}
	if err := a.Redis.Validate(); err != nil {
		return err
	}
	if a.DatabaseUrl == "" {
		return errEmptyDatabaseUrl
	}

	return nil
}

func (a *App) Version() string {
	return a.version
}

func (a *App) RunMode() string {
	return a.runMode
}

func (a *App) loadEnvs() error {
	if a.DatabaseUrl = os.Getenv("CCDC_DATABASEURL"); a.DatabaseUrl == "" {
		return fmt.Errorf("failed to load 'CCDC_DATABASEURL' env: %w", errEmptyDatabaseUrl)
	}

	if dataProvider := os.Getenv("CCDC_DATAPROVIDER"); dataProvider != "" {
		a.DataProvider = strings.ToLower(dataProvider)
	}

	a.ApiKey = os.Getenv("CCDC_APIKEY")

	if sessionStore := os.Getenv("CCDC_SESSIONSTORE"); sessionStore != "" {
		a.SessionStore = strings.ToLower(sessionStore)
	}

	if strings.ToLower(os.Getenv("CCDC_DEBUG")) == "true" {
		a.debug = true
	}
	if a.debug {
		a.runMode = gin.DebugMode
	}

	return nil
}
