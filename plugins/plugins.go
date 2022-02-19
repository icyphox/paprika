package plugins

import (
	"errors"
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
	Execute(command string, rest string, m *irc.Message) (*irc.Message, error)
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
	Matches(m *irc.Message) (string, error)
}

var matchers = []MatchPlugin{}

func RegisterMatcher(p MatchPlugin) {
	matchers = append(matchers, p)
}

// This Error is used to signal a NAK so other plugins
// can attempt to parse the result
// This is useful for broad prefix matching or future regexp
// functions, other special errors may need defining
// to determin priority or to "concatenate" output.
var NoReply = errors.New("No Reply")

// Due to racey nature of the handler in main.go being invoked as a goroutine,
// it's hard to have race free way of building correct state of the IRC world.
// This is a temporary (lol) hack to check if the PRIVMSG target looks like a
// IRC nickname. We assume that any IRC nickname target would be us, unless
// the IRCd we are on is neurotic
//
// Normally one would use the ISUPPORT details to learn what prefixes are used
// on the network's valid channel names.
//
// E.G. ISUPPORT ... CHANTYPES=# ... where # would be the only valid channel name
// allowed on the IRCd.
func unlikelyDirectMessage(target string) bool {
	if len(target) < 1 {
		panic("Conformity Error, IRCd sent us a PRIVMSG with an empty target and message.")
	}

	sym := target[0] // we only care about the byte (ASCII)
	return likelyInvalidNickChr(sym)
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
func ProcessTrigger(m *irc.Message) ([]*irc.Message, error) {
	// ignore anyone with a "bot" like name if configured.
	if config.C.IgnoreBots && strings.HasSuffix(strings.ToLower(m.Name), "bot") {
		return nil, NoReply
	}
	if !unlikelyDirectMessage(m.Params[0]) {
		m.Params[0] = m.Name
	}

	var replies []*irc.Message

	for _, matcher := range matchers {
		match, err := matcher.Matches(m)
		var reply *irc.Message
		if err == nil {
			reply, err = matcher.Execute(match, m.Trailing(), m)
		}

		if err == NoReply {
			continue
		} else if err != nil {
			return nil, err
		}

		replies = append(replies, reply)
	}

	command_rest := strings.SplitN(m.Trailing(), " ", 2)
	command := command_rest[0]
	var message string
	if len(command_rest) == 2 && command != "" {
		message = command_rest[1]
	}

	if plug, ok := simplePlugins[command]; ok {
		reply, err := plug.Execute(command, message, m)
		if err == NoReply {
		} else if err != nil {
			return nil, err
		} else {
			replies = append(replies, reply)
		}
	}

	if len(replies) == 0 {
		return nil, NoReply
	} else {
		return replies, nil
	}
}
