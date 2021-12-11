package plugins

import (
	"errors"
	"strconv"
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

// Checks for triggers in a message and executes its
// corresponding plugin, returning the response/error.
func ProcessTrigger(m *irc.Message) (string, error) {
	if !unlikelyDirectMessage(m.Params[0]) {
		m.Params[0] = m.Name
	}

	for trigger, plugin := range Plugins {
		if strings.HasPrefix(m.Trailing(), trigger) {
			response, err := plugin.Execute(m)
			if !errors.Is(err, NoReply) {
				return response, err
			}
		}
	}
	return "", NoReply // No plugin matched, so we need to Ignore.
}

func IntersperseThousandsSepInt(num int, sep rune) string {
	numstr := strconv.Itoa(num)
	acc := ""

	i := 0
	for j := len(numstr) - 1; j > -1; j-- {
		if i != 0 && i % 3 == 0 {
			acc = "," + acc
		}
		acc = string(numstr[j]) + acc
		i++
	}
	return acc
}
