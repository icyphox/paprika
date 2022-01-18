package plugins

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/irc.v3"
)

type Plugin interface {
	Triggers() []string
	Execute(m *irc.Message) (string, error)
}

type PluginTuple struct {
	a string
	b Plugin
}

var Plugins = []PluginTuple{}

func Register(p Plugin) {
	for _, t := range p.Triggers() {
		Plugins = append(Plugins, PluginTuple{a: t, b: p})
	}
}

// This Error is used to signal a NAK so other plugins
// can attempt to parse the result
// This is useful for broad prefix matching or future regexp
// functions, other special errors may need defining
// to determin priority or to "concatenate" output.
var NoReply = errors.New("No Reply")

// This error indicates we are sending a NOTICE instead of a PRIVMSG
var IsNotice = errors.New("Is Notice")

// This means the string(s) we are returning are raw IRC commands
// that need to be written verbatim.
var IsRaw = errors.New("Is Raw")

// This error indicates that the message is a NOTICE, along with a
// recepient.
type IsPrivateNotice struct {
	To string
}

func (e *IsPrivateNotice) Error() string {
	return fmt.Sprintf("Is Private Notice: %s", e.To)
}

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
func ProcessTrigger(m *irc.Message) (string, error) {
	if !unlikelyDirectMessage(m.Params[0]) {
		m.Params[0] = m.Name
	}

	for _, pluginTup := range Plugins {
		if strings.HasPrefix(m.Trailing(), pluginTup.a) {
			response, err := pluginTup.b.Execute(m)
			if !errors.Is(err, NoReply) {
				return response, err
			}
		}
	}
	return "", NoReply // No plugin matched, so we need to Ignore.
}
