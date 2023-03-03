# ccd

[![Go Report Card](https://goreportcard.com/badge/github.com/streamdp/ccd)](https://goreportcard.com/report/github.com/streamdp/ccd)
[![Website ccdtest.gq](https://img.shields.io/website-up-down-green-red/https/ccdtest.gq.svg)](https://ccdtest.gq/)
[![GitHub release](https://img.shields.io/github/release/streamdp/ccd.svg)](https://github.com/streamdp/ccd/releases/)
[![GitHub license](https://img.shields.io/github/license/streamdp/ccd.svg)](https://github.com/streamdp/ccd/blob/main/LICENSE)

It is a microservice that collect data from a several dataproviders using its API.

This microservice uses:

* gin-gonic/gin package to start and serve HTTP server
* gorilla/websocket package to work through websockets
* go-sql-driver/mysql package to work with mysql database

## Build app

```bash
$ go build -o ccd .
````

## Run app
You should previously export some environment variables:

```bash
export CCDC_DATAPROVIDER=cryptocompare
export CCDC_DATASOURCE=mysql://username:password@tcp(localhost:3306)/dbname
export CCDC_APIKEY=put you api key here
```
And run application:
```bash
$ ./ccd -debug
```

The default port is 8080, you can test the application in a browser or with curl:

```bash
$ curl 127.0.0.1:8080/v1/service/ping
```

You can choose a different port and run more than one copy of **ccd** on your local host. For example:

```bash
$ ./ccd -port 8081
``` 

You also can specify some setting before run application: 
```bash
$ ./ccd -h
ccd is a microservice that collect data from a cryprocompare using its API.

Usage of ccd:
  -common string
        specify list possible common currencies (default "USD,EUR,GBP,JPY,RUR")
  -crypto string
        specify list possible crypto currencies (default "BTC,XRP,ETH,BCH,EOS,LTC,XMR,DASH")
  -dataprovider string
        use selected data provider ("cryptocompare", "huobi") (default "cryptocompare")
  -debug
        run the program in debug mode
  -h    display help
  -port string
        set specify port (default ":8080")
  -timeout int
        how long to wait for a response from the api server before sending data from the cache (default 1000)
```

List of the implemented endpoints:
* **/v1/service/ping** [GET]   _check node status_
* **/v1/collect/add** [POST, GET] _add new worker to collect data about selected pair in database_
* **/v1/collect/remove** [POST, GET] _stop and remove worker and collecting data for selected pair_
* **/v1/collect/status** [GET] _show info about running workers_
* **/v1/collect/update** [POST, GET]  _update pulling interval for selected pair_
* **/v1/price** [POST, GET] _get actual (or cached if dataprovider is unavailable) info for selected pair_
* **/v1/ws** [GET] _websocket connection url_

Example getting a GET request for getting actual info about selected pair:

```bash
$ curl "http://localhost:8080/v1/price?fsym=ETH&tsym=JPY"
```

Example of sending a POST request to add a new worker:

```bash
$ curl -X POST -H "Content-Type: application/json" -d '{ "fsym": "BTC", "tsym": "USD", "interval": 60}' "http://localhost:8080/v1/collect/add"
```

Example of sending a GET request to remove worker:

```bash
$ curl "http://localhost:8080/v1/collect/remove?fsym=BTC&tsym=USD&interval=60"
```

Working example URL: https://ccdtest.gq/v1/service/ping

Web UI: https://ccdtest.gq
