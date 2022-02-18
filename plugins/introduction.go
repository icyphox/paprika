package plugins

import (
	"fmt"

	"git.icyphox.sh/paprika/database"
	"github.com/dgraph-io/badger/v3"
	"gopkg.in/irc.v3"
)

func init() {
	Register(Introduction{})
}

type Introduction struct{}

func (Introduction) Triggers() []string {
	return []string{".intro"}
}

func (Introduction) Execute(cmd, rest string, m *irc.Message) (*irc.Message, error) {
	userKey := database.ToKey("introduction", m.Name)
	if rest == "" {
		intro, err := database.DB.Get(userKey)
		if err == badger.ErrKeyNotFound {
			return NewRes(m, fmt.Sprintf("Usage: %s <intro text>", cmd)), nil
		} else if err != nil {
			return nil, err
		} else {
			return NewRes(m, string(intro)), nil
		}
	}

	err := database.DB.Set(userKey, []byte(rest))
	if err != nil {
		return nil, err
	}
	return nil, NoReply
}

func GetIntro(m *irc.Message) (*irc.Message, error) {
	intro, err := database.DB.Get(database.ToKey("introduction", m.Name))
	if err == badger.ErrKeyNotFound {
		return nil, NoReply
	} else if err != nil {
		return nil, err
	} else {
		return &irc.Message{
			Command: "PRIVMSG",
			Params:  []string{m.Params[0], string(intro)},
		}, nil
	}
}
