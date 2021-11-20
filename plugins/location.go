package plugins

import (
	"fmt"
	"strings"

	"gopkg.in/irc.v3"
)

func init() {
	Register(Location{})
}

type Location struct{}

func (Location) Triggers() []string {
	return []string{".loc", ".location"}
}

func (Location) Execute(m *irc.Message) (string, error) {
	parsed := strings.SplitN(m.Trailing(), " ", 2)
	trigger := parsed[0]
	location := parsed[1]
	if len(parsed) != 2 {
		return fmt.Sprintf("Usage: %s <location>", trigger), nil
	}
}
