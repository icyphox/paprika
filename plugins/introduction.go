package plugins

import (
	"fmt"
	"log"

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

func (Introduction) Execute(cmd, rest string, c *irc.Client, m *irc.Message) {
	userKey := database.ToKey("introduction", m.Name)
	if rest == "" {
		intro, err := database.DB.Get(userKey)
		if err == badger.ErrKeyNotFound {
			c.WriteMessage(NewRes(m, fmt.Sprintf("Usage: %s <intro text>", cmd)))
			return
		} else if err != nil {
			log.Println(err)
			return
		} else {
			c.WriteMessage(NewRes(m, string(intro)))
			return
		}
	}

	err := database.DB.Set(userKey, []byte(rest))
	if err != nil {
		log.Println(err)
		return
	}
}

func GetIntro(c *irc.Client, m *irc.Message) {
	intro, err := database.DB.Get(database.ToKey("introduction", m.Name))
	if err == badger.ErrKeyNotFound {
	} else if err != nil {
		log.Println(err)
	} else {
		c.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params:  []string{m.Params[0], string(intro)},
		})
	}
}
