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
	return []string{"." + string(n), ".kiss"}
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
	case ".kiss", ".love":
		rand.Seed(time.Now().Unix())
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
	}
	return "", nil
}
