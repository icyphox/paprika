package plugins

import (
	"fmt"
	"log"
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

func (Version) Execute(cmd, rest string, c *irc.Client, m *irc.Message) {
	if rest == "" {
		c.WriteMessage(NewRes(m, fmt.Sprintf("Usage: %s <nick>", cmd)))
	}

	nickTarget := rest
	if likelyInvalidNick(nickTarget) {
		c.WriteMessage(NewRes(m, fmt.Sprintf("%s does not look like an IRC Nick", nickTarget)))
	}

	nickKey := database.ToKey("version", nickTarget)
	replyTarget := m.Params[0]
	err := database.DB.SetWithTTL(
		nickKey,
		[]byte(replyTarget),
		2*time.Minute,
	)
	if err != nil {
		log.Println(err)
		return
	}

	m.Params[0] = nickTarget
	c.WriteMessage(&irc.Message{
		Command: "PRIVMSG",
		Params:  []string{nickTarget, "\x01VERSION\x01"},
	})
}

func CTCPReply(c *irc.Client, m *irc.Message) {
	msg := m.Trailing()
	if !strings.HasPrefix(msg, "\x01VERSION") {
		return
	}

	params := strings.SplitN(msg, " ", 2)
	if len(params) != 2 {
		return
	}
	if len(params[1]) == 0 {
		return
	}

	ver := params[1][:len(params[1])-1]
	from := m.Name

	newTarget, err := database.DB.Delete(database.ToKey("version", from))
	if err != nil {
		log.Println(err)
	}

	c.WriteMessage(&irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			string(newTarget),
			fmt.Sprintf("%s VERSION: %s", from, ver),
		},
	})
}

func NoSuchUser(c *irc.Client, m *irc.Message) {
	invalidNick := m.Params[1]
	newTarget, err := database.DB.Delete(database.ToKey("version", invalidNick))
	if err == badger.ErrKeyNotFound {
		return
	} else if err != nil {
		log.Println(err)
		return
	}

	c.WriteMessage(&irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			string(newTarget),
			fmt.Sprintf("No such nickname: %s", invalidNick),
		},
	})
}
