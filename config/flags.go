package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

const DefaultPullingInterval = 60

var Version = "1.0.0"

// LoadConfig and update config variables
func LoadConfig() (*App, error) {
	var (
		showHelp    bool
		showVersion bool
		debug       bool
		appCfg      = NewAppConfig()
	)
	flag.BoolVar(&showHelp, "h", false, "display help")
	flag.BoolVar(&showVersion, "v", false, "display version")
	flag.BoolVar(&debug, "debug", false, "run the program in debug mode")
	flag.IntVar(&appCfg.Http.port, "port", httpServerDefaultPort, "set specify port")
	flag.StringVar(&appCfg.SessionStore, "session", defaultSessionStore,
		"set session store \"db\" or \"redis\"")
	flag.IntVar(&appCfg.Http.clientTimeout, "timeout", httpDefaultTimeout, "how long to wait for a "+
		"response from the api server before sending data from the cache")
	flag.StringVar(&appCfg.DataProvider, "dataprovider", defaultDataProvider, "use selected data provider"+
		" (\"cryptocompare\", \"huobi\")")
	flag.Parse()

	if appCfg.DatabaseUrl = os.Getenv("CCDC_DATABASEURL"); appCfg.DatabaseUrl == "" {
		return nil, errEmptyDatabaseUrl
	}
	if os.Getenv("CCDC_DEBUG") != "" {
		debug = true
	}
	if dataProvider := os.Getenv("CCDC_DATAPROVIDER"); dataProvider != "" {
		appCfg.DataProvider = strings.ToLower(dataProvider)
	}
	if sessionStore := os.Getenv("CCDC_SESSIONSTORE"); sessionStore != "" {
		appCfg.SessionStore = strings.ToLower(sessionStore)
	}
	appCfg.ApiKey = os.Getenv("CCDC_APIKEY")
	if showHelp {
		fmt.Println("ccd is a microservice that collect data from several crypto data providers using its API.")
		fmt.Println("")
		flag.Usage()
		os.Exit(1)
	}
	if showVersion {
		fmt.Println("ccd version: " + Version)
		os.Exit(1)
	}

	appCfg.Version = Version
	appCfg.RunMode = gin.ReleaseMode
	if debug {
		appCfg.RunMode = gin.DebugMode
	}

	if err := appCfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid app config: %w", err)
	}

	return appCfg, nil
}
