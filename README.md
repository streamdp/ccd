# CCD
[![Website ccd.oncook.top](https://img.shields.io/website-up-down-green-red/https/ccd.oncook.top/healthz.svg)](https://ccd.oncook.top/)
[![GitHub release](https://img.shields.io/github/release/streamdp/ccd.svg)](https://github.com/streamdp/ccd/releases/)

This microservice is designed to manage real-time and historical cryptocurrency data collection. It provides both REST 
and WebSocket endpoints for flexible interaction with currency pair data. The service supports the following key 
functionalities:
* **Multiple Data Provider Support**: You can choose between **cryptocompare**, **huobi** or **kraken** platform to 
collect market data.
* **Worker Management**: You can add, update, list, or remove background workers responsible for collecting data for 
specific currency pairs. These workers handle data pulling at defined intervals.
* **Symbol Management**: Add, update, list, or delete currency symbols that are tracked by the system.
* **Price Retrieval**: Fetch the most recent price data for a selected currency pair, including cached data when the 
data provider is unavailable.
* **WebSocket Integration**: Provides a WebSocket connection to receiving real-time data. Clients can receive market info 
for specific currency pairs using JSON messages like _{"fsym":"BTC","tsym":"USD"}_.
## Build app
```bash
$ go build -o ccd .
````
## Run app
To configure app, export some environment variables:
```bash
export CCDC_DATAPROVIDER=cryptocompare #huobi, kraken
export CCDC_DATABASEURL=postgres://username:password@127.0.0.1:5432/dbname?sslmode=disable
export CCDC_APIKEY=put you api key here
export CCDC_SESSIONSTORE=redis // or "db", default value is "db"
export REDIS_URL=redis://:redis_password@127.0.0.1:6379/0 // only when "redis" session store selected
```
And run application:
```bash
$ ./ccd -debug
```
The default port is 8080, test the application in a browser or with curl:
```bash
$ curl 127.0.0.1:8080/healthz
```
Choose a different port and run more than one copy of **ccd** on the local host. For example:
```bash
$ ./ccd -port 8081
``` 
Specify some setting before run application: 
```bash
$ ./ccd -h
ccd is a microservice that collect data from several crypto data providers cryprocompare using its API.

Usage of ccd:
  -dataprovider string
        use selected data provider ("cryptocompare", "huobi", "kraken") (default "cryptocompare")
  -debug
        run the program in debug mode
  -h    display help
  -port int
        set specify port (default 8080)
  -session string
        set session store "db" or "redis" (default "db")
  -timeout int
        how long to wait for a response from the api server before sending data from the cache (default 5000)
  -v    display version

```
Since the release of v2.3.0, the ccd service has moved to API v2, all v1 endpoints have been deprecated and 
are not recommended for use. List of the implemented endpoints v2 API:

| Method | Endpoint               | Description                                                                                        |
|:------:|:-----------------------|:---------------------------------------------------------------------------------------------------|
|  GET   | **/healthz**           | check node status                                                                                  |
|  GET   | **/v2/collect**        | list of all running workers                                                                        |
|  POST  | **/v2/collect**        | add new worker to collect data for the selected pair                                               |
|  PUT   | **/v2/collect**        | update pulling interval for the selected pair                                                      |
| DELETE | **/v2/collect**        | stop and remove worker and collecting data for the selected pair                                   |
|  GET   | **/v2/symbols**        | list of all symbols presented                                                                      |
|  POST  | **/v2/symbols**        | add currency symbol                                                                                |
|  PUT   | **/v2/symbols**        | update currency symbol                                                                             |
| DELETE | **/v2/symbols**        | delete currency symbol                                                                             |
|  GET   | **/v2/price**          | get actual (or cached when dataprovider is unavailable) info for the selected pair                 |
|  GET   | **/v2/ws**             | websocket connection url, when you connected, try to send request like {"fsym":"BTC","tsym":"USD"} |
|  GET   | **/v2/ws/subscribe**   | subscribe to collect data for the selected pair                                                    |
|  GET   | **/v2/ws/unsubscribe** | unsubscribe to stop collect data for the selected pair                                             |
## Usage examples
Get actual info about selected pair:
```bash
$ curl "http://localhost:8080/v2/price?fsym=ETH&tsym=USDT"
```
Add a new worker:
```bash
$ curl -X POST -H "Content-Type: application/json" -d '{ "fsym": "BTC", "tsym": "USD", "interval": 60}' "http://localhost:8080/v2/collect"
```
Remove worker:
```bash
$ curl -X DELETE "http://localhost:8080/v2/collect?fsym=BTC&tsym=USD&interval=60"
```
Subscribe wss channel:
```bash
$ curl "http://localhost:8080/v2/ws/subscribe?fsym=BTC&tsym=USD"
```
Web UI: https://ccd.oncook.top
## Contributing
Contributions are welcome! If you encounter any issues, have suggestions for new features, or want to improve **CCD**, please feel free to open an issue or submit a pull request on the project's GitHub repository.
## License
**CCD** is released under the _MIT_ License. See the [LICENSE](https://github.com/streamdp/ccd/blob/main/LICENSE) file for complete license details.
