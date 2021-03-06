package plugins

import (
	"fmt"
	"log"
	"strings"

	coingecko "git.icyphox.sh/paprika/plugins/coingecko"
	"github.com/dustin/go-humanize"
	"gopkg.in/irc.v3"
)

type CoinGecko struct{}

func init() {
	Register(CoinGecko{})
}

var (
	twoArg   = []string{".cg", ".coin", ".crypto"}
	aliases  = []string{".btc", ".eth", ".doge", ".bnb"}
	triggers = append(twoArg, aliases...)
)

func (CoinGecko) Triggers() []string {
	return triggers
}

func formatCgNum(field string, value float64, percent bool) string {
	if percent {
		v := humanize.CommafWithDigits(value+0.00000000001, 2)
		if value < 0 {
			return fmt.Sprintf("%s: \x0304%s%%\x03 ", field, v)
		} else {
			return fmt.Sprintf("%s: \x0303%s%%\x03 ", field, v)
		}
	} else if value >= 0.01 {
		v := humanize.CommafWithDigits(value+0.00000000001, 2)
		return fmt.Sprintf("%s: $%s ", field, v)
	} else {
		return fmt.Sprintf("%s: $%.3e ", field, value)
	}
}

func (CoinGecko) Execute(cmd, rest string, c *irc.Client, m *irc.Message) {
	var coin string
	if rest == "" {
		for _, twoarg := range twoArg {
			if cmd == twoarg {
				c.WriteMessage(NewRes(m, fmt.Sprintf("Usage: %s <Ticker>", cmd)))
				return
			}
		}
		for _, alias := range aliases {
			if cmd == alias {
				coin = alias[1:]
				break
			}
		}
	} else {
		coin = rest
	}

	err := coingecko.CheckUpdateCoinList()
	if err != nil {
		log.Println(err)
		return
	}

	cid, err := coingecko.GetCoinId(coin)
	if err != nil {
		log.Println(err)
		return
	} else if cid == "" {
		c.WriteMessage(NewRes(m, fmt.Sprintf("No such coin found: %s", coin)))
		return
	}

	stats, err := coingecko.GetCoinPrice(cid)
	if err != nil {
		log.Println(err)
		return
	} else if stats == nil {
		c.WriteMessage(NewRes(m, fmt.Sprintf("No such coin found: %s", coin)))
		return
	}

	mData := stats.MarketData

	var outRes strings.Builder
	outRes.WriteString(fmt.Sprintf(
		"\x02%s (%s)\x02 - ",
		stats.Name, stats.Symbol,
	))
	outRes.WriteString(formatCgNum("Current", mData.Current.Usd, false))
	outRes.WriteString(formatCgNum("High", mData.High.Usd, false))
	outRes.WriteString(formatCgNum("Low", mData.Low.Usd, false))

	outRes.WriteString(formatCgNum("MCap", mData.MarketCap.Usd, false))
	outRes.WriteString(fmt.Sprintf("(#%s) ", humanize.Comma(mData.MarketCapRank)))

	outRes.WriteString(formatCgNum("24h", mData.Change24h, true))
	outRes.WriteString(formatCgNum("7d", mData.Change7d, true))
	outRes.WriteString(formatCgNum("30d", mData.Change30d, true))
	outRes.WriteString(formatCgNum("60d", mData.Change60d, true))
	outRes.WriteString(formatCgNum("1y", mData.Change1y, true))

	c.WriteMessage(NewRes(m, outRes.String()))
	return
}
