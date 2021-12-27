package plugins

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"gopkg.in/irc.v3"
)

func init() {
	Register(Seen{})
}

type Seen struct{}

func (Seen) Triggers() []string {
	return []string{".seen", ""}
}

var LastSeen sync.Map

func (Seen) Execute(m *irc.Message) (string, error) {
	// we just saw this user so.
	LastSeen.Store(m.Name, time.Now())

	if strings.HasPrefix(m.Params[1], ".seen") {
		params := strings.Split(m.Trailing(), " ")
		if len(params) == 1 {
			return ".seen nickname", nil
		}

		if seen, ok := LastSeen.Load(params[1]); ok {
			return fmt.Sprintf(
				"\x02%s\x02 was last seen: %s",
				params[1], humanize.Time(seen.(time.Time)),
			), nil
		} else {
			return "I have not seen " + params[1], nil
		}
	}

	return "", NoReply
}
