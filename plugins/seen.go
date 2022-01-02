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

type LastSeenInfo struct {
	// The last message the user sent.
	Message string
	// The last time this user was seen.
	Time time.Time
}

func (Seen) Execute(m *irc.Message) (string, error) {
	LastSeen.Store(m.Name, LastSeenInfo{
		Message: m.Trailing(),
		// We just saw the user, so.
		Time: time.Now(),
	})

	if strings.HasPrefix(m.Params[1], ".seen") {
		params := strings.Split(m.Trailing(), " ")
		if len(params) == 1 {
			return "Usage: .seen <nickname>", nil
		}

		if seen, ok := LastSeen.Load(params[1]); ok {
			humanized := humanize.Time(seen.(LastSeenInfo).Time)

			// Don't want "now ago".
			if humanized != "now" {
				humanized = humanized + " ago"
			}

			return fmt.Sprintf(
				"\x02%s\x02 was last seen %s, saying: %s",
				params[1], humanized,
				seen.(LastSeenInfo).Message,
			), nil
		} else {
			return "I have not seen " + params[1], nil
		}
	}

	return "", NoReply
}
