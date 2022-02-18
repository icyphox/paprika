package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	stockClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	api_endpoint = "https://cloud.iexapis.com/v1"
	NoIEXApi     = errors.New("No IEX API key")
)

func (Stocks) Triggers() []string {
	return []string{".stock", ".stonk"}
}

type tickerData struct {
	Quote struct {
		Symbol        string  `json:"symbol"`
		Current       float64 `json:"latestPrice"`
		High          float64 `json:"high,omitempty"`
		Low           float64 `json:"low,omitempty"`
		ChangePercent float64 `json:"changePercent"`
	} `json:"quote"`
	Stats struct {
		Company   string  `json:"companyName"`
		Change1y  float64 `json:"year1ChangePercent"`
		Change6M  float64 `json:"month6ChangePercent"`
		Change30d float64 `json:"day30ChangePercent"`
		Change5d  float64 `json:"day5ChangePercent"`
	} `json:"stats"`
}

func formatMoneyNum(field string, value float64, percent bool) string {
	if percent {
		v := humanize.CommafWithDigits(value*100+0.00000000001, 2)
		if value < 0 {
			return fmt.Sprintf("%s: \x0304%s%%\x03 ", field, v)
		} else {
			return fmt.Sprintf("%s: \x0303%s%%\x03 ", field, v)
		}
	} else {
		v := humanize.CommafWithDigits(value+0.00000000001, 2)
		return fmt.Sprintf("%s: $%s ", field, v)
	}
}

func getStock(symbol, apiKey string) (string, error) {
	req, err := http.NewRequest("GET", api_endpoint+"/stock/market/batch", nil)
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
	outRes.WriteString(formatMoneyNum("Current", quote.Current, false))
	if quote.High != 0.0 {
		outRes.WriteString(formatMoneyNum("High", quote.High, false))
		outRes.WriteString(formatMoneyNum("Low", quote.Low, false))
	}
	outRes.WriteString(formatMoneyNum("24h", quote.ChangePercent, true))
	outRes.WriteString(formatMoneyNum("5d", stats.Change5d, true))
	outRes.WriteString(formatMoneyNum("6M", stats.Change6M, true))
	outRes.WriteString(formatMoneyNum("1y", stats.Change1y, true))

	return outRes.String(), nil
}

func (Stocks) Execute(cmd, rest string, m *irc.Message) (*irc.Message, error) {
	if rest == "" {
		return NewRes(m, fmt.Sprintf("Usage: %s <Ticker>", rest)), nil
	}
	sym := strings.ToUpper(rest)

	if apiKey, ok := config.C.ApiKeys["iex"]; ok {
		res, err := getStock(sym, apiKey)
		return NewRes(m, res), err
	}

	return nil, NoIEXApi
}
