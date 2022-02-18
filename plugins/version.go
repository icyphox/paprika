package plugins

import (
	"fmt"
	"strings"
	"time"

	"git.icyphox.sh/paprika/database"
	"github.com/dgraph-io/badger/v3"
	"gopkg.in/irc.v3"
)

func init() {
	Register(Version{})
}

type Version struct{}

func (Version) Triggers() []string {
	return []string{".v", ".version"}
}

func (Version) Execute(cmd, rest string, m *irc.Message) (*irc.Message, error) {
	if rest == "" {
		return NewRes(m, fmt.Sprintf("Usage: %s <nick>", cmd)), nil
	}

	nickTarget := rest
	if likelyInvalidNick(nickTarget) {
		return NewRes(m, fmt.Sprintf("%s does not look like an IRC Nick", nickTarget)), nil
	}

	nickKey := database.ToKey("version", nickTarget)
	replyTarget := m.Params[0]
	err := database.DB.SetWithTTL(
		nickKey,
		[]byte(replyTarget),
		2*time.Minute,
	)
	if err != nil {
		return nil, err
	}

	m.Params[0] = nickTarget
	return &irc.Message{
		Command: "PRIVMSG",
		Params:  []string{nickTarget, "\x01VERSION\x01"},
	}, nil
}

func CTCPReply(m *irc.Message) (*irc.Message, error) {
	msg := m.Trailing()
	if !strings.HasPrefix(msg, "\x01VERSION") {
		return nil, NoReply
	}

	params := strings.SplitN(msg, " ", 2)
	if len(params) != 2 {
		return nil, NoReply
	}
	if len(params[1]) == 0 {
		return nil, NoReply
	}

	ver := params[1][:len(params[1])-1]
	from := m.Name

	newTarget, err := database.DB.Delete(database.ToKey("version", from))
	if err != nil {
		return nil, err
	}

	return &irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			string(newTarget),
			fmt.Sprintf("%s VERSION: %s", from, ver),
		},
	}, nil
}

func NoSuchUser(m *irc.Message) (*irc.Message, error) {
	invalidNick := m.Params[1]
	newTarget, err := database.DB.Delete(database.ToKey("version", invalidNick))
	if err == badger.ErrKeyNotFound {
		return nil, NoReply
	} else if err != nil {
		return nil, err
	}

	return &irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			string(newTarget),
			fmt.Sprintf("No such nickname: %s", invalidNick),
		},
	}, nil
}
