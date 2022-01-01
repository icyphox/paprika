package plugins

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

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

func (WolframAlpha) Execute(m *irc.Message) (string, error) {
	parts := strings.SplitN(m.Trailing(), " ", 2)
	trigger := parts[0]
	if len(parts) < 2 {
		return fmt.Sprintf("Usage: %s <query>", trigger), nil
	}
	query := url.QueryEscape(parts[1])

	appID := config.C.ApiKeys["wolframalpha"]
	url := fmt.Sprintf(
		"https://api.wolframalpha.com/v1/result?i=%s&appid=%s",
		query, appID,
	)

	r, err := http.Get(url)
	if err != nil || r.StatusCode != 200 {
		return "Error getting result", err
	}

	result, err := io.ReadAll(r.Body)
	if err != nil {
		return "Error getting result", err
	}
	return fmt.Sprintf("\x02Result:\x02 %s", string(result)), nil
}
