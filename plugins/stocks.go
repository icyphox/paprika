package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"git.icyphox.sh/paprika/config"
	"github.com/dustin/go-humanize"
	"gopkg.in/irc.v3"
)

func init() {
	Register(Stocks{})
}

type Stocks struct{}

var (
	aliases = []string{".btc", ".eth"}
	triggers = append([]string{".stock", ".stonk", ".crypto", ".cmc", ".cg"}, aliases...)
	stockClient = &http.Client {
		Timeout: 10 * time.Second,
	}
	api_endpoint = "https://cloud.iexapis.com/v1"
    NoIEXApi = errors.New("No IEX API key")
)

func (Stocks) Triggers() []string {
	return triggers
}

type tickerData struct{
	Quote struct {
		Symbol string `json:"symbol"`
		Current float64 `json:"latestPrice"`
		High float64 `json:"high,omitempty"`
		Low float64 `json:"low,omitempty"`
		ChangePercent float64 `json:"changePercent"`
	} `json:"quote"`
	Stats struct {
		Company string `json:"companyName"`
		Change1y float64 `json:"year1ChangePercent"`
		Change6M float64 `json:"month6ChangePercent"`
		Change30d float64 `json:"day30ChangePercent"`
		Change5d float64 `json:"day5ChangePercent"`
	} `json:"stats"`
}

type cryptoData struct {
	Price float64 `json:"price,string"`
	Symbol string `json:"symbol"`
}

func formatNum(field string, value float64, percent bool) string {
	if percent {
		v := humanize.CommafWithDigits(value * 100 + 0.00000000001, 2)
		if value < 0 {
			return fmt.Sprintf("%s: \x0304%s%%\x03 ", field, v)
		} else {
			return fmt.Sprintf("%s: \x0303%s%%\x03 ", field, v)
		}
	} else {
		v := humanize.CommafWithDigits(value + 0.00000000001, 2)
		return fmt.Sprintf("%s: $%s ", field, v)
	}
}

func getCrypto(symbol, apiKey string) (string, error) {
	endpoint := fmt.Sprintf("%s/crypto/%s/price", api_endpoint, url.PathEscape(symbol))
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "[Crypto] Request construction error.", err
	}
	req.Header.Add("User-Agent", "github.com/icyphox/paprika")
	q := req.URL.Query()
	q.Add("token", apiKey)
	req.URL.RawQuery = q.Encode()

	res, err := stockClient.Do(req)
	if err != nil {
		return "[Crypto] API Client Error", err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return fmt.Sprintf("[Crypto] Could not get quote for \x02%s\x02", strings.ToUpper(symbol)), nil
	}

	var resData cryptoData
	err = json.NewDecoder(res.Body).Decode(&resData)
	if err != nil {
		return "[Crypto] API response malformed", err
	}

	return fmt.Sprintf("\x02%s\x02 Price: $%s", resData.Symbol, humanize.CommafWithDigits(resData.Price + 0.00000001, 2)), nil
}

func getStock(symbol, apiKey string) (string, error) {
	req, err := http.NewRequest("GET", api_endpoint + "/stock/market/batch", nil)
	if err != nil {
		return "[Stocks] Request construction error.", err
	}
	req.Header.Add("User-Agent", "github.com/icyphox/paprika")
	q := req.URL.Query()
	q.Add("token", apiKey)
	q.Add("symbols", symbol)
	q.Add("types", "quote,stats")
	req.URL.RawQuery = q.Encode()

	res, err := stockClient.Do(req)
	if err != nil {
		return "[Stocks] API Client Error", err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return fmt.Sprintf("[Stocks] Could not get quote for \x02%s\x02", symbol), nil
	}

	var resData map[string]tickerData
	err = json.NewDecoder(res.Body).Decode(&resData)
	if err != nil {
		return "[Stock] API response malformed", err
	}

	quote := resData[symbol].Quote
	stats := resData[symbol].Stats
	var outRes strings.Builder
	outRes.WriteString(fmt.Sprintf("\x02%s (%s)\x02 - ", stats.Company, quote.Symbol))
	outRes.WriteString(formatNum("Current", quote.Current, false))
	if quote.High != 0.0 {
		outRes.WriteString(formatNum("High", quote.High, false))
		outRes.WriteString(formatNum("Low", quote.Low, false))
	}
	outRes.WriteString(formatNum("24h", quote.ChangePercent, true))
	outRes.WriteString(formatNum("5d", stats.Change5d, true))
	outRes.WriteString(formatNum("6M", stats.Change6M, true))
	outRes.WriteString(formatNum("1y", stats.Change1y, true))

	return outRes.String(), nil
}

func (Stocks) Execute(m *irc.Message) (string, error) {
	parsed := strings.SplitN(m.Trailing(), " ", 3)
	if len(parsed) > 2 {
		return fmt.Sprintf("Usage: %s <Ticker>", parsed[0]), nil
	}
	cmd := parsed[0]

	// crypto shortcuts
	coin := ""
	for _, a := range aliases {
		if cmd == a {
			coin = strings.ToLower(cmd[1:]) + "usd"
		}
	}

	var sym string
	if len(parsed) != 2 && coin == "" {
		return fmt.Sprintf("Usage: %s <Ticker>", parsed[0]), nil
	} else if coin != "" {
		sym = coin
	} else {
		sym = strings.ToUpper(parsed[1])
	}

	if apiKey, ok := config.C.ApiKeys["iex"]; ok {
		switch cmd {
		case ".stock", ".stonk":
			return getStock(sym, apiKey)
		default:
			return getCrypto(sym, apiKey)
		}
	}

	return "", NoIEXApi
}
