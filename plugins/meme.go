package plugins

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"gopkg.in/irc.v3"
)

func init() {
	Register(Meme{})
}

type Meme struct{}

var n []byte

func (Meme) Triggers() []string {
	n, _ = base64.StdEncoding.DecodeString("bmlnZ2Vy")
	return []string{
		"." + string(n),
		".kiss",
		".increase",
		".decrease",
		".sniff",
		".hug",
	}
}

func (Meme) Execute(m *irc.Message) (string, error) {
	parts := strings.SplitN(m.Trailing(), " ", 2)
	trigger := parts[0]
	rand.Seed(time.Now().Unix())
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
		word := string(n)
		if rand.Intn(10) == 8 {
			// Easter egg! Only teh cool h4x0rz will get this.
			word = "bmlnZ2Vy"
		}
		return fmt.Sprintf("%s is a %s", target, word), nil
	case ".kiss", ".love":
		kaomoji := []string{
			"(●´□`)", "(｡･ω･｡)ﾉ", "(｡’▽’｡)",
			"(ෆ ͒•∘̬• ͒)◞", "( •ॢ◡-ॢ)-", "⁽⁽ପ( •ु﹃ •ु)​.⑅*",
			"(๑ Ỡ ◡͐ Ỡ๑)ﾉ", "◟(◔ั₀◔ั )◞ ༘",
		}
		return fmt.Sprintf(
			"%s \x02\x034 。。・゜゜・。。・❤️ %s ❤️ \x03\x02",
			kaomoji[rand.Intn(len(kaomoji))],
			target,
		), nil
	case ".increase", ".decrease":
		return fmt.Sprintf(
			"\x02[QUALITY OF CHANNEL SIGNIFICANTLY %sD]\x02",
			strings.ToUpper(trigger[1:]),
		), nil
	case ".sniff":
		return fmt.Sprintf("huffs %s's hair while sat behind them on the bus.", target), nil
	case ".hug":
		return fmt.Sprintf("(>^_^)>❤️ %s ❤️<(^o^<)", target), nil

	}
	return "", nil
}
