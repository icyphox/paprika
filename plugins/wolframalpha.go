package plugins

import (
	"fmt"
	"io"
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

func (WolframAlpha) Execute(cmd, rest string, m *irc.Message) (*irc.Message, error) {
	if rest == "" {
		return NewRes(m, fmt.Sprintf("Usage: %s <query>", cmd)), nil
	}
	query := url.QueryEscape(rest)

	if appID, ok := config.C.ApiKeys["wolframalpha"]; ok {
		url := fmt.Sprintf(
			"https://api.wolframalpha.com/v1/result?i=%s&appid=%s",
			query, appID,
		)

		r, err := http.Get(url)
		if err != nil || r.StatusCode != 200 {
			return nil, err
		}

		result, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		return NewRes(m, fmt.Sprintf("\x02Result:\x02 %s", string(result))), nil
	}

	return NewRes(m, "Plugin Disabled"), nil
}
