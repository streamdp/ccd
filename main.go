package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/streamdp/ccdatacollector/dataproviders"
	"github.com/streamdp/ccdatacollector/dataproviders/cryptocompare"
	"github.com/streamdp/ccdatacollector/dbconnectors"
	"github.com/streamdp/ccdatacollector/utility"
	"strconv"
)

func main() {
	var ccData *cryptocompare.CryptoCompare
	var showHelp bool
	var port int
	var d dbconnectors.Db
	ccData.SetApiKey(utility.GetEnv("CCDC_APIKEY"))
	ccData.SetApiURL(utility.GetEnv("CCDC_APIURL"))
	if ccData.GetApiURL() == "" || ccData.GetApiKey() == "" {
		utility.HandleError(errors.New("you should specify \"CCDC_APIKEY\" and \"CCDC_APIURL\" in you OS environment"))
		return
	}
	dataSource := utility.GetEnv("CCDC_DATASOURCE")
	if dataSource == "" {
		utility.HandleError(errors.New("please set OS environment \"CCDC_DATASOURCE\" with database connection string"))
		return
	}
	flag.BoolVar(&showHelp, "h", false, "display help")
	flag.IntVar(&port, "port", 8080, "set specify port")
	flag.Parse()
	if showHelp {
		fmt.Println("ccdatacollector is a microservice that collect data from a cryprocompare using its API.")
		fmt.Println("")
		flag.Usage()
		return
	}
	if err := d.Connect(dataSource); err != nil {
		utility.HandleError(err)
		return
	}
	defer func(d *dbconnectors.Db) {
		if err := d.Close(); err != nil {
			utility.HandleError(err)
		}
	}(&d)
	wc := dataproviders.CreateWorkersControl()
	pipe := make(chan *dataproviders.DataPipe, 20)
	defer close(pipe)
	go dbconnectors.ServePipe(pipe, &d)
	r := gin.Default()
	if err := r.SetTrustedProxies(nil); err != nil {
		utility.HandleError(err)
	}
	r.GET("/ping", ping)
	r.POST("/collect/add", collectAdd(ccData, wc, pipe))
	r.POST("/collect/remove", collectRemove(wc))
	r.GET("/collect/status", collectStatus(wc))

	if err := r.Run(":" + strconv.Itoa(port)); err != nil {
		return
	}
}
