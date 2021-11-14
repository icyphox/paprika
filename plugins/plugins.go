package plugins

import (
	"strings"

	"gopkg.in/irc.v3"
)

type Plugin interface {
	Triggers() []string
	Execute(m *irc.Message) (string, error)
}

var Plugins = make(map[string]Plugin)

func Register(p Plugin) {
	for _, t := range p.Triggers() {
		Plugins[t] = p
	}
}

// Checks for triggers in a message and executes its
// corresponding plugin, returning the response/error.
func ProcessTrigger(m *irc.Message) (string, error) {
	var (
		response string
		err      error
	)
	for trigger, plugin := range Plugins {
		if strings.HasPrefix(m.Trailing(), trigger) {
			response, err = plugin.Execute(m)
			if err != nil {
				return "", err
			}
			return response, nil
		}
	}
	return "", nil
}
