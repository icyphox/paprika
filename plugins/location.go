package plugins

import (
	"fmt"
	"strings"

	"git.icyphox.sh/paprika/plugins/location"
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
	if len(parsed) != 2 {
		return fmt.Sprintf("Usage: %s <location>", trigger), nil
	}
	loc := parsed[1]

	err := location.SetLocation(loc, m.Prefix.Name)
	if err != nil {
		return "Error setting location", err
	}

	return "Successfully set location", nil
}
