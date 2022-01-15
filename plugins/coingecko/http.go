package coingecko

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

var (
	CoinGeckoClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	apiEndpoint = "https://api.coingecko.com/api/v3/coins"
)

type CoinId struct {
	Id     string `json:"id"`
	Symbol string `json:"symbol"`
}

type CoinData struct {
	Name       string `json:"name"`
	Symbol     string `json:"symbol"`
	MarketData struct {
		Current struct {
			Usd float64 `json:"usd"`
		} `json:"current_price"`
		High struct {
			Usd float64 `json:"usd"`
		} `json:"high_24h"`
		Low struct {
			Usd float64 `json:"usd"`
		} `json:"low_24h"`
		MarketCap struct {
			Usd float64 `json:"usd"`
		} `json:"market_cap"`
		MarketCapRank int64   `json:"market_cap_rank"`
		Change24h     float64 `json:"price_change_percentage_24h"`
		Change7d      float64 `json:"price_change_percentage_7d"`
		Change14d     float64 `json:"price_change_percentage_14d"`
		Change30d     float64 `json:"price_change_percentage_30d"`
		Change60d     float64 `json:"price_change_percentage_60d"`
		Change200d    float64 `json:"price_change_percentage_200d"`
		Change1y      float64 `json:"price_change_percentage_1y"`
	} `json:"market_data"`
}

func GetCoinList() ([]CoinId, error) {
	req, err := http.NewRequest("GET", apiEndpoint+"/list", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "github.com/icyphox/paprika")

	res, err := CoinGeckoClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var resData []CoinId
	err = json.NewDecoder(res.Body).Decode(&resData)
	if err != nil {
		return nil, err
	}

	return resData, nil
}

func GetCoinPrice(coinId string) (*CoinData, error) {
	cid := url.PathEscape(coinId)
	req, err := http.NewRequest("GET", apiEndpoint+"/"+cid, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "github.com/icyphox/paprika")

	res, err := CoinGeckoClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return nil, nil
	}

	var resData *CoinData
	err = json.NewDecoder(res.Body).Decode(&resData)
	if err != nil {
		return nil, err
	}

	return resData, nil
}
