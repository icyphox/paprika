package plugins

import (
	"fmt"
	"strings"

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

func (Introduction) Execute(m *irc.Message) (string, error) {
	param := strings.SplitN(m.Trailing(), " ", 2)
	userKey := database.ToKey("introduction", m.Name)
	if len(param) != 2 {
		intro, err := database.DB.Get(userKey)
		if err == badger.ErrKeyNotFound {
			return fmt.Sprintf("Usage: %s <intro text>", param[0]), nil
		} else if err != nil {
			return "Unknown Error Checking for your intro.", err
		} else {
			return string(intro), nil
		}
	}

	err := database.DB.Set(userKey, []byte(param[1]))
	if err != nil {
		return "[Introduction] Failed to set introduction string.", nil
	}

	return "", NoReply
}

func GetIntro (m *irc.Message) (string, error) {
	intro, err := database.DB.Get(database.ToKey("introduction", m.Name))
	if err == badger.ErrKeyNotFound {
		return "", NoReply
	} else if err != nil {
		return "", err
	} else {
		return string(intro), nil
	}
}
