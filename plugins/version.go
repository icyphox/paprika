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

func (Version) Execute(m *irc.Message) (string, error) {
	params := strings.SplitN(m.Trailing(), " ", 3)
	if len(params) != 2 {
		return fmt.Sprintf("Usage: %s <nick>", params[0]), nil
	}

	nickTarget := params[1]
	if likelyInvalidNick(nickTarget) {
		return fmt.Sprintf("%s does not look like an IRC Nick", nickTarget), nil
	}

	nickKey := database.ToKey("version", nickTarget)
	replyTarget := m.Params[0]
	err := database.DB.SetWithTTL(
		nickKey,
		[]byte(replyTarget),
		2 * time.Minute,
	)
	if err != nil {
		return "", err
	}

	m.Params[0] = nickTarget
	return "\x01VERSION\x01", nil
}

func CTCPReply(m *irc.Message) (string, error) {
	msg := m.Trailing()
	if !strings.HasPrefix(msg, "\x01VERSION") {
		return "", NoReply
	}

	params := strings.SplitN(msg, " ", 2)
	if len(params) != 2 {
		return "", NoReply
	}
	if len(params[1]) == 0 {
		return "", NoReply
	}

	ver := params[1][:len(params[1])-1]
	from := m.Name

	newTarget, err := database.DB.Delete(database.ToKey("version", from))
	if err != nil {
		return "", err
	}
	
	m.Params[0] = string(newTarget)
	m.Command = "PRIVMSG"
	return fmt.Sprintf("%s VERSION: %s", from, ver), nil
}

func NoSuchUser(m *irc.Message) (string, error) {
	invalidNick := m.Params[1]
	newTarget, err := database.DB.Delete(database.ToKey("version", invalidNick))
	if err == badger.ErrKeyNotFound {
		return "", NoReply
	} else if err != nil {
		return "", err
	}

	m.Command = "PRIVMSG"
	m.Params[0] = string(newTarget)
	m.Params = m.Params[:1]
	return fmt.Sprintf("No such nickname: %s", invalidNick), nil
}
