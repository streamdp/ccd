package config

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"strings"
)

var (
	Port              = ":8080"
	RunMode           = gin.DebugMode
	Crypto            = "BTC,XRP,ETH,BCH,EOS,LTC,XMR,DASH"
	Common            = "USD,EUR,GBP,JPY,RUR"
	HttpClientTimeout = 1000
)

func ParseFlags() {
	var showHelp bool
	var debug bool
	flag.BoolVar(&showHelp, "h", false, "display help")
	flag.BoolVar(&debug, "debug", false, "run the program in debug mode")
	flag.StringVar(&Port, "port", ":8080", "set specify port")
	flag.IntVar(&HttpClientTimeout, "timeout", HttpClientTimeout, "how long to wait for a response from the"+
		" api server before sending data from the cache")
	flag.StringVar(&Common, "common", Common, "specify list possible common currencies")
	flag.StringVar(&Crypto, "crypto", Crypto, "specify list possible crypto currencies")
	flag.Parse()
	if showHelp {
		fmt.Println("ccdatacollector is a microservice that collect data from a cryprocompare using its API.")
		fmt.Println("")
		flag.Usage()
		os.Exit(1)
	}
	if !debug {
		RunMode = gin.ReleaseMode
	}
}

func GetEnv(name string) (result string) {
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if pair[0] == name {
			return pair[1]
		}
	}
	return
}
