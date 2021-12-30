package plugins

import (
	"encoding/base64"
	"fmt"
	"strings"

	"gopkg.in/irc.v3"
)

func init() {
	Register(Meme{})
}

type Meme struct{}

var n []byte

func (Meme) Triggers() []string {
	n, _ = base64.StdEncoding.DecodeString("bmlnZ2Vy")
	return []string{"." + string(n)}
}

func (Meme) Execute(m *irc.Message) (string, error) {
	parts := strings.SplitN(m.Trailing(), " ", 2)
	trigger := parts[0]
	var target string
	if len(parts) > 1 {
		target = parts[1]
	} else {
		target = m.Prefix.Name
	}

	switch trigger {
	case "." + string(n):
		// lol
		if m.Prefix.Name == "IRSSucks" {
			target = "IRSSucks"
		}
		return fmt.Sprintf("%s is a %s", target, string(n)), nil
	}
	return "", nil
}
