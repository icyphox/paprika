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

type Flag int

type ReplyT struct {
	Flags Flag
}

const (
	Notice Flag = 1 << iota
	DirectMessage
)

var AllFlags = [...]Flag{Notice, DirectMessage}

func (f Flag) String() string {
	switch f {
	case Notice:
		return "Notice"
	case DirectMessage:
		return "DM"
	default:
		return fmt.Sprintf("Invalid:%d", f)
	}
}

func NewReplyT(flags Flag) *ReplyT {
	return &ReplyT{Flags: flags}
}

func (r *ReplyT) ApplyFlags(m *irc.Message) {
	for _, flag := range AllFlags {
		if r.Flags & flag != 0 {
			switch flag {
			case Notice:
				m.Command = "NOTICE"
			case DirectMessage:
				m.Params[0] = m.Name
			}
		}
	}
}

func (r *ReplyT) Error() string {
	var reply strings.Builder
	for _, flag := range AllFlags {
		if r.Flags & flag != 0 {
			reply.WriteString(flag.String())
		}
	}
	return reply.String()
}

func IsReplyT(e error) bool {
	r := &ReplyT{}
	return errors.As(e, &r)
}

// Checks for triggers in a message and executes its
// corresponding plugin, returning the response/error.
func ProcessTrigger(m *irc.Message) (string, error) {
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
