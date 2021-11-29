package ws

import (
	"log"
)

type DataHub struct {
	cccAgg     chan *CccAgg
	heartBeat  chan *HeartBeat
	register   chan *Client
	unregister chan *Client
	clients    map[*Client]bool
}

type HeartBeat struct {
	Type    string `json:"TYPE"`
	Message string `json:"MESSAGE"`
	Timems  int64  `json:"TIMEMS"`
}

type CccAgg struct {
	Circulatingsupply       int     `json:"CIRCULATINGSUPPLY"`
	Circulatingsupplymktcap float64 `json:"CIRCULATINGSUPPLYMKTCAP"`
	Currentsupply           int     `json:"CURRENTSUPPLY"`
	Currentsupplymktcap     float64 `json:"CURRENTSUPPLYMKTCAP"`
	Fromsymbol              string  `json:"FROMSYMBOL"`
	High24Hour              float64 `json:"HIGH24HOUR"`
	Highday                 float64 `json:"HIGHDAY"`
	Highhour                float64 `json:"HIGHHOUR"`
	Lastmarket              string  `json:"LASTMARKET"`
	Lasttradeid             string  `json:"LASTTRADEID"`
	Lastupdate              int     `json:"LASTUPDATE"`
	Lastvolume              float64 `json:"LASTVOLUME"`
	Lastvolumeto            float64 `json:"LASTVOLUMETO"`
	Low24Hour               float64 `json:"LOW24HOUR"`
	Lowday                  float64 `json:"LOWDAY"`
	Lowhour                 float64 `json:"LOWHOUR"`
	Market                  string  `json:"MARKET"`
	Maxsupply               float64 `json:"MAXSUPPLY"`
	Maxsupplymktcap         float64 `json:"MAXSUPPLYMKTCAP"`
	Median                  float64 `json:"MEDIAN"`
	Open24Hour              float64 `json:"OPEN24HOUR"`
	Openday                 float64 `json:"OPENDAY"`
	Openhour                float64 `json:"OPENHOUR"`
	Price                   float64 `json:"PRICE"`
	Toptiervolume24Hour     float64 `json:"TOPTIERVOLUME24HOUR"`
	Toptiervolume24Hourto   float64 `json:"TOPTIERVOLUME24HOURTO"`
	Tosymbol                string  `json:"TOSYMBOL"`
	Type                    string  `json:"TYPE"`
	Volume24Hour            float64 `json:"VOLUME24HOUR"`
	Volume24Hourto          float64 `json:"VOLUME24HOURTO"`
	Volumeday               float64 `json:"VOLUMEDAY"`
	Volumedayto             float64 `json:"VOLUMEDAYTO"`
	Volumehour              float64 `json:"VOLUMEHOUR"`
	Volumehourto            float64 `json:"VOLUMEHOURTO"`
}

func NewHub() *DataHub {
	return &DataHub{
		cccAgg:     make(chan *CccAgg),
		heartBeat:  make(chan *HeartBeat),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *DataHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case cccAgg := <-h.cccAgg:
			log.Println(cccAgg)
		case heartBeat := <-h.heartBeat:
			log.Println(heartBeat)
		}
	}
}
