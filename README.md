# ccd

It is a microservice that collect data from a cryprocompare using its API.

This microservice uses:

* gin-gonic/gin package to start and serve HTTP server
* gorilla/websocket to work through websockets
* go-sql-driver/mysql to work with mysql database

## Build the app

```bash
$ go build -o ccd .
````

## Run the app

```bash
$ ./ccd
```

The default port is 8080, you can test the application in a browser or with curl:

```bash
$ curl 127.0.0.1:8080/v1/service/ping
```

You can choose a different port and run more than one copy of **ccd** on your local host. For example:

```bash
$ ./ccd -port 8081
``` 

List of the endpoints:

* GET  **/v1/service/ping** _check node status_

* POST, GET **/v1/collect/add** _add new worker to collect data about selected pair in database_
* POST, GET **/v1/collect/remove** _stop and remove worker and collecting data for selected pair_
* GET **/v1/collect/status** _show info about running workers_
* POST, GET **/v1/collect/update** _update pulling interval for selected pair_
* POST, PRICE **/v1/price** _get actual (or cached if dataprovider is unavailable) info for selected pair_
* WS **/v1/ws** _websocket connection url_

Example getting a GET request for getting actual info about selected pair:

```bash
$ curl "https://ccdtest.gq/v1/price?fsym=ETH&tsym=JPY"
```

Example of sending a POST request to add a new worker:

```bash
$ curl -X POST -H "Content-Type: application/json" -d '{ "fsym": "BTC", "tsym": "USD", "interval": 60}' "http://localhost:8080/v1/collect/add"
```

Example of sending a GET request to remove worker:

```bash
$ curl "https://localhost:8080/v1/collect/remove?fsym=BTC&tsym=USD&interval=60"
```

Working example URL: https://ccdtest.gq/v1/service/ping
