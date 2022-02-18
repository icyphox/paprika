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

func (Ctcp) Execute(cmd, rest string, m *irc.Message) (*irc.Message, error) {
	msg := m.Trailing()
	reply := &irc.Message{
		Tags:    nil,
		Prefix:  nil,
		Command: "NOTICE",
		Params:  []string{m.Params[0], ""},
	}

	if msg == "\x01VERSION\x01" {
		reply.Params[1] = "\x01VERSION git.icyphox.sh/paprika\x01"
	} else if strings.HasPrefix(msg, "\x01PING") {
		reply.Params[1] = msg
	}

	return reply, nil
}
