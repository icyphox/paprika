// Sanity check plugin.
package plugins

import (
	"strings"

	"gopkg.in/irc.v3"
)

func init() {
	Register(Hello{})
	RegisterMatcher(Hello{})
}

type Hello struct{}

func (Hello) Triggers() []string {
	return []string{".hello"}
}

func (Hello) Matches(c *irc.Client, m *irc.Message) (string, bool) {
	if strings.Contains(m.Trailing(), c.CurrentNick()) {
		return m.Prefix.Name, true
	}
	return "", false
}

func (Hello) Execute(matched, rest string, c *irc.Client, m *irc.Message) {
	c.WriteMessage(NewRes(m, "hello, "+matched))
}
