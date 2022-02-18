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

func (Hello) Matches(m *irc.Message) (string, error) {
	if strings.HasPrefix(m.Trailing(), "paprika") {
		return "paprika", nil
	}

	return "", NoReply
}

func (Hello) Execute(cmd, rest string, m *irc.Message) (*irc.Message, error) {
	return NewRes(m, "hello, "+m.Prefix.Name), nil
}
