package market

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
)

type (
	gate struct{}
	gateTickerResponse struct {
		Last decimal.Decimal `json:"last"`
	}
)

func (ex *gate) GetPrice() (price decimal.Decimal, err error) {
	url := "https://data.gateio.la/api2/1/ticker/nax_usdt"
	resp, err := http.Get(url)
	if err != nil {
		return price, fmt.Errorf("http.Get: %s", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return price, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return price, fmt.Errorf("ioutil.ReadAll: %s", err.Error())
	}
	var ticker gateTickerResponse
	err = json.Unmarshal(data, &ticker)
	if err != nil {
		return price, fmt.Errorf("json.Unmarshal: %s", err.Error())
	}
	price = ticker.Last
	if price.IsZero() {
		return price, fmt.Errorf("invalid response")
	}
	return price, nil
}

