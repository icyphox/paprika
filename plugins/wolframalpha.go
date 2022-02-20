package plugins

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"git.icyphox.sh/paprika/config"
	"gopkg.in/irc.v3"
)

func init() {
	Register(WolframAlpha{})
}

type WolframAlpha struct{}

func (WolframAlpha) Triggers() []string {
	return []string{".wa", ".calc"}
}

func (WolframAlpha) Execute(cmd, rest string, c *irc.Client, m *irc.Message) {
	if rest == "" {
		c.WriteMessage(NewRes(m, fmt.Sprintf("Usage: %s <query>", cmd)))
		return
	}
	query := url.QueryEscape(rest)

	if appID, ok := config.C.ApiKeys["wolframalpha"]; ok {
		url := fmt.Sprintf(
			"https://api.wolframalpha.com/v1/result?i=%s&appid=%s",
			query, appID,
		)

		r, err := http.Get(url)
		if err != nil {
			log.Println(err)
			return
		} else if r.StatusCode != 200 {
			log.Println("We got a bad reply from Wolfram, check API key.")
			return
		}

		result, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
		} else {
			c.WriteMessage(NewRes(m, fmt.Sprintf("\x02Result:\x02 %s", string(result))))
		}
	} else {
		c.WriteMessage(NewRes(m, "Plugin Disabled (No API Key)"))
	}
}
