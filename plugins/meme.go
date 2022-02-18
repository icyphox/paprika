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

func (Meme) Execute(cmd, rest string, m *irc.Message) (*irc.Message, error) {
	rand.Seed(time.Now().Unix())
	var target string
	if rest != "" {
		target = rest
	} else {
		target = m.Prefix.Name
	}

	var response string
	switch cmd {
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
		response = fmt.Sprintf("%s is a %s", target, word)
	case ".kiss", ".love":
		kaomoji := []string{
			"(●´□`)", "(｡･ω･｡)ﾉ", "(｡’▽’｡)",
			"(ෆ ͒•∘̬• ͒)◞", "( •ॢ◡-ॢ)-", "⁽⁽ପ( •ु﹃ •ु)​.⑅*",
			"(๑ Ỡ ◡͐ Ỡ๑)ﾉ", "◟(◔ั₀◔ั )◞ ༘",
		}
		response = fmt.Sprintf(
			"%s \x02\x034 。。・゜゜・。。・❤️ %s ❤️ \x03\x02",
			kaomoji[rand.Intn(len(kaomoji))],
			target,
		)
	case ".increase", ".decrease":
		response = fmt.Sprintf(
			"\x02[QUALITY OF CHANNEL SIGNIFICANTLY %sD]\x02",
			strings.ToUpper(cmd[1:]),
		)
	case ".sniff":
		response = fmt.Sprintf("huffs %s's hair while sat behind them on the bus.", target)
	case ".hug":
		response = fmt.Sprintf("(>^_^)>❤️ %s ❤️<(^o^<)", target)
	default:
		panic("Unreachable!")
	}

	return NewRes(m, response), nil
}
