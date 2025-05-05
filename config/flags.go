package config

import (
	"flag"
	"fmt"
	"os"
)

const DefaultPullingInterval = 60

var version = "1.0.0"

// LoadConfig and update config variables
func LoadConfig() (*App, error) {
	var (
		showHelp    bool
		showVersion bool
		appCfg      = NewAppConfig()
	)
	flag.BoolVar(&showHelp, "h", false, "display help")
	flag.BoolVar(&showVersion, "v", false, "display version")
	flag.BoolVar(&appCfg.debug, "debug", false, "run the program in debug mode")
	flag.IntVar(&appCfg.Http.port, "port", httpServerDefaultPort, "set specify port")
	flag.StringVar(&appCfg.SessionStore, "session", defaultSessionStore,
		"set session store \"db\" or \"redis\"")
	flag.IntVar(&appCfg.Http.clientTimeout, "timeout", httpDefaultTimeout, "how long to wait for a "+
		"response from the api server before sending data from the cache")
	flag.StringVar(&appCfg.DataProvider, "dataprovider", defaultDataProvider, "use selected data provider"+
		" (\"cryptocompare\", \"huobi\", \"kraken\")")
	flag.Parse()

	if showHelp {
		fmt.Println("ccd is a microservice that collect data from several crypto data providers using its API.")
		fmt.Println("")
		flag.Usage()
		os.Exit(1)
	}
	if showVersion {
		fmt.Println("ccd version: " + version)
		os.Exit(1)
	}

	if err := appCfg.loadEnvs(); err != nil {
		return nil, fmt.Errorf("failed to load os envs: %w", err)
	}

	if err := appCfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid app config: %w", err)
	}

	return appCfg, nil
}
