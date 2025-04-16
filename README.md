# CCD
[![Website ccd.oncook.top](https://img.shields.io/website-up-down-green-red/https/ccd.oncook.top/healthz.svg)](https://ccd.oncook.top/)
[![GitHub release](https://img.shields.io/github/release/streamdp/ccd.svg)](https://github.com/streamdp/ccd/releases/)

This microservice is designed to manage real-time and historical cryptocurrency data collection. It provides both RESTful and WebSocket endpoints for flexible interaction with currency pair data. The service supports the following key functionalities:
* **Worker Management**: You can add, update, list, or remove background workers responsible for collecting data for specific currency pairs. These workers handle data pulling at defined intervals.
* **Symbol Management**: Add, update, list, or delete currency symbols that are tracked by the system.
* **Price Retrieval**: Fetch the most recent price data for a selected currency pair, including cached data when the data provider is unavailable.
* **WebSocket Integration**: Provides a WebSocket connection for real-time updates. Clients can subscribe or unsubscribe to specific currency pairs using JSON messages like _{"fsym":"BTC","tsym":"USD"}_.
## Build app

```bash
$ go build -o ccd .
````

## Run app
To configure app, export some environment variables:
```bash
export CCDC_DATAPROVIDER=cryptocompare #huobi
export CCDC_DATABASEURL=postgres://username:password@127.0.0.1:5432/dbname?sslmode=disable
export CCDC_APIKEY=put you api key here
export CCDC_SESSIONSTORE=redis // or "db", default value is "db"
export REDIS_URL=redis://:redis_password@127.0.0.1:6379/0 // only when "redis" session store selected
```
To use **mysql** db, just export something like this instead:
```bash
export CCDC_DATABASEURL=mysql://username:password@tcp(localhost:3306)/dbname
``` 
And run application:
```bash
$ ./ccd -debug
```

The default port is 8080, you can test the application in a browser or with curl:

```bash
$ curl 127.0.0.1:8080/healthz
```

You can choose a different port and run more than one copy of **ccd** on your local host. For example:

```bash
$ ./ccd -port 8081
``` 

You also can specify some setting before run application: 
```bash
$ ./ccd -h
ccd is a microservice that collect data from several crypto data providers cryprocompare using its API.

Usage of ccd:
  -dataprovider string
        use selected data provider ("cryptocompare", "huobi") (default "cryptocompare")
  -debug
        run the program in debug mode
  -h    display help
  -port string
        set specify port (default ":8080")
  -session string
        set session store "db" or "redis" (default "db")  
  -timeout int
        how long to wait for a response from the api server before sending data from the cache (default 1000)
```
Since the release of v2.3.0, the ccd service has moved to API v2, all v1 endpoints have been deprecated and 
are not recommended for use. List of the implemented endpoints v2 API:
* **/healthz** [GET]   _check node status_
* **/v2/collect** [GET] _list of all running workers_
* **/v2/collect** [POST] _add new worker to collect data for the selected pair_
* **/v2/collect** [PUT]  _update pulling interval for the selected pair_
* **/v2/collect** [DELETE] _stop and remove worker and collecting data for the selected pair_
* **/v2/symbols** [GET] _list of all symbols presented_
* **/v2/symbols** [POST] _add currency symbol_
* **/v2/symbols** [PUT] _update currency symbol_
* **/v2/symbols** [DELETE] _delete currency symbol_
* **/v2/price** [GET] _get actual (or cached when dataprovider is unavailable) info for the selected pair_
* **/v2/ws** [GET] _websocket connection url, when you connected, try to send request like {"fsym":"BTC","tsym":"USD"}_
* **/v2/ws/subscribe** [GET] _subscribe to collect data for the selected pair_
* **/v2/ws/unsubscribe** [GET] _unsubscribe to stop collect data for the selected pair_

Example getting a GET request for getting actual info about selected pair:

```bash
$ curl "http://localhost:8080/v2/price?fsym=ETH&tsym=USDT"
```

Example of sending a POST request to add a new worker:

```bash
$ curl -X POST -H "Content-Type: application/json" -d '{ "fsym": "BTC", "tsym": "USD", "interval": 60}' "http://localhost:8080/v2/collect"
```

Example of sending a DELETE request to remove worker:

```bash
$ curl -X DELETE "http://localhost:8080/v2/collect?fsym=BTC&tsym=USD&interval=60"
```

Example of sending a GET request to subscribe wss channel:

```bash
$ curl "http://localhost:8080/v2/ws/subscribe?fsym=BTC&tsym=USD"
```

Working example URL: https://ccd.oncook.top/healthz

Web UI: https://ccd.oncook.top

## Contributing
Contributions are welcome! If you encounter any issues, have suggestions for new features, or want to improve **CCD**, please feel free to open an issue or submit a pull request on the project's GitHub repository.
## License
**CCD** is released under the _MIT_ License. See the [LICENSE](https://github.com/streamdp/ccd/blob/main/LICENSE) file for complete license details.
## Support project
[DigitalOcean](https://www.digitalocean.com/?refcode=253bf19488bd&utm_campaign=Referral_Invite&utm_medium=Referral_Program) referral link.
