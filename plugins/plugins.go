package plugins

import (
	"strings"

	"git.icyphox.sh/paprika/config"
	"gopkg.in/irc.v3"
)

func NewRes(m *irc.Message, reply string) *irc.Message {
	return &irc.Message{
		Tags:    nil,
		Prefix:  nil,
		Command: "PRIVMSG",
		Params: []string{
			m.Params[0],
			reply,
		},
	}
}

type Plugin interface {
	Triggers() []string
	Execute(command string, rest string, c *irc.Client, m *irc.Message)
}

var simplePlugins = make(map[string]Plugin)

func Register(p Plugin) {
	for _, t := range p.Triggers() {
		simplePlugins[t] = p
	}
}

// A match plugin is a plugin that has no specific command but reacts
// To a specific match in an IRC message.
// Example would be a url matcher that would fetch youtube metadata
// when a user posts a url like https://youtu.be/<VID_ID_HERE>.
type MatchPlugin interface {
	Plugin
	// Execute's command will be the matching value returned.
	// If the returned "match string" is empty, it is assumed to not match.
	Matches(c *irc.Client, m *irc.Message) (string, bool)
}

var matchers = []MatchPlugin{}

func RegisterMatcher(p MatchPlugin) {
	matchers = append(matchers, p)
}

func likelyInvalidNickChr(sym byte) bool {
	// Is one of: !"#$%&'()*+,_./
	// or one of: ;<=>?@
	// If your IRCd uses symbols outside of this range,
	// god help us.
	//
	// ALSO NOTE: RFC1459 defines a "nick" as
	// <nick> ::= <letter> { <letter> | <number> | <special> }
	// But I have seen some networks that allow special/number as the first letter.
	return sym > 32 /* SPACE */ && sym < 48 /* 0 */ ||
		sym > 58 /* : */ && sym < 65 /* A */ ||
		sym == 126 /* ~ */
}

func likelyInvalidNick(nick string) bool {
	for i := 0; i < len(nick); i++ {
		if likelyInvalidNickChr(nick[i]) {
			return true
		}
	}
	return false
}

// Checks for triggers in a message and executes its
// corresponding plugin, returning the response/error.
func ProcessTrigger(c *irc.Client, m *irc.Message) {
	// ignore anyone with a "bot" like name if configured.
	if config.C.IgnoreBots && strings.HasSuffix(strings.ToLower(m.Name), "bot") {
		return
	}
	if !c.FromChannel(m) {
		m.Params[0] = m.Name
	}

	for _, matcher := range matchers {
		if match, ok := matcher.Matches(c, m); ok {
			matcher.Execute(match, m.Trailing(), c, m)
		}
	}

	command_rest := strings.SplitN(m.Trailing(), " ", 2)
	command := command_rest[0]
	var message string
	if len(command_rest) == 2 && command != "" {
		message = command_rest[1]
	}

	if plug, ok := simplePlugins[command]; ok {
		plug.Execute(command, message, c, m)
	}
}
