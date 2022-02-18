package plugins

import (
	"fmt"

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

func (Location) Execute(cmd, rest string, m *irc.Message) (*irc.Message, error) {
	if rest == "" {
		return NewRes(m, fmt.Sprintf("Usage: %s <location>", cmd)), nil
	}
	loc := rest

	err := location.SetLocation(loc, m.Prefix.Name)
	if err != nil {
		return nil, err
	}

	return NewRes(m, "Successfully set location"), nil
}
