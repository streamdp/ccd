package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

const DefaultPullingInterval = 60

var (
	// Port - set default port for gin-gonic engine init
	Port              = ":8080"
	RunMode           = gin.DebugMode
	Crypto            = "BTC,XRP,ETH,BCH,EOS,LTC,XMR,DASH"
	Common            = "USD,EUR,GBP,JPY,RUR"
	HttpClientTimeout = 1000
	Version           = "1.0.0"
	DataProvider      = "cryptocompare" // "huobi"
)

// ParseFlags and update config variables
func ParseFlags() {
	var (
		showHelp    bool
		showVersion bool
		debug       bool
	)
	flag.BoolVar(&showHelp, "h", false, "display help")
	flag.BoolVar(&showVersion, "v", false, "display version")
	flag.BoolVar(&debug, "debug", false, "run the program in debug mode")
	flag.StringVar(&Port, "port", ":8080", "set specify port")
	flag.IntVar(&HttpClientTimeout, "timeout", HttpClientTimeout, "how long to wait for a response from the"+
		" api server before sending data from the cache")
	flag.StringVar(&Common, "common", Common, "specify list possible common currencies")
	flag.StringVar(&Crypto, "crypto", Crypto, "specify list possible crypto currencies")
	flag.StringVar(&DataProvider, "dataprovider", DataProvider, "use selected data provider"+
		" (\"cryptocompare\", \"huobi\")")
	flag.Parse()
	if GetEnv("CCDC_DEBUG") != "" {
		debug = true
	}
	if dataProvider := GetEnv("CCDC_DATAPROVIDER"); dataProvider != "" {
		DataProvider = dataProvider
	}
	if common := GetEnv("CCDC_COMMON"); common != "" {
		Common = common
	}
	if crypto := GetEnv("CCDC_CRYPTO"); crypto != "" {
		Crypto = crypto
	}
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
	if !debug {
		RunMode = gin.ReleaseMode
	}
}

// GetEnv values for selected name or return null
func GetEnv(name string) (result string) {
	return localEnvs.get(name)
}
