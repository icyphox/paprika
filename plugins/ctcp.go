package plugins

import (
	"strings"

	"gopkg.in/irc.v3"
)

type Ctcp struct{}

func init() {
	Register(Ctcp{})
}

func (Ctcp) Triggers() []string {
	return []string{"\x01VERSION\x01", "\x01PING"}
}

func (Ctcp) Execute(m *irc.Message) (string, error) {
	msg := m.Trailing()
	if msg == "\x01VERSION\x01" {
		return "\x01VERSION git.icyphox.sh/paprika\x01", IsNotice
	} else if strings.HasPrefix(msg, "\x01PING") {
		return msg, IsNotice
	}

	panic("Unreachable!")
}
