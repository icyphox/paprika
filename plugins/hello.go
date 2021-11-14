// Sanity check plugin.
package plugins

import (
	"gopkg.in/irc.v3"
)

func init() {
	Register(Hello{})
}

type Hello struct{}

func (Hello) Triggers() []string {
	return []string{".hello", "taigobot"}
}

func (Hello) Execute(m *irc.Message) (string, error) {
	return "hello, " + m.Prefix.Name, nil
}
