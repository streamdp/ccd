package dataproviders

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Response struct {
		Change24Hour    float64 `json:"CHANGE24HOUR"`
		Changepct24Hour float64 `json:"CHANGEPCT24HOUR"`
		Open24Hour      float64 `json:"OPEN24HOUR"`
		Volume24Hour    float64 `json:"VOLUME24HOUR"`
		Volume24Hourto  float64 `json:"VOLUME24HOURTO"`
		Low24Hour       float64 `json:"LOW24HOUR"`
		High24Hour      float64 `json:"HIGH24HOUR"`
		Price           float64 `json:"PRICE"`
		Supply          float64 `json:"SUPPLY"`
		Mktcap          float64 `json:"MKTCAP"`
		Lastupdate      int     `json:"LASTUPDATE"`
}

type Data struct {
	Raw map[string]map[string]Response `json:"RAW"`
}

type DataPipe struct {
	From string
	To string
	Data *Data
}

type DataProvider interface {
	GetSerializable() *Data
	GetApiURLToCompare(fSym string, tSym string) (*url.URL, error)
}

func PullingData(dp DataProvider, from string, to string, interval int, worker *Worker) (err error){
	var (
		data     []byte
		response *http.Response
		apiURL   *url.URL
	)
	defer func(w *Worker) {
		w.SetAlive(false)
		close(w.GetDone())
	}(worker)
	worker.SetAlive(true)
	if apiURL, err = dp.GetApiURLToCompare(from, to); err != nil {
		return errors.Wrapf(err, "failed to get url")
	}
	for {
		select {
		case <-worker.GetDone():
			return
		case <-time.After(time.Duration(interval) * time.Second):
			if response, err = http.Get(apiURL.String()); err != nil {
				return errors.Wrapf(err, "failed to open URL")
			}
			if data, err = ioutil.ReadAll(response.Body); err != nil {
				return errors.Wrapf(err,"failed to read response body")
			}
			if response.StatusCode == 200 {
				ds := dp.GetSerializable()
				if err = json.Unmarshal(data, &ds); err != nil {
					return errors.Wrapf(err, "failed to unmarshal data")
				}
				worker.GetPipe() <- &DataPipe{
					From:from,
					To: to,
					Data: ds,
				}
			} else {
				return errors.Wrapf(errors.New("server error:"), "error getting data with %v statuscode", response.StatusCode)
			}
		}
	}
}

func GetEmptyData(from string, to string) *Data {
	result := &Data{
		Raw: map[string]map[string]Response{},
	}
	result.Raw[from] = make(map[string]Response)
	result.Raw[from][to] = Response{}
	return result
}
