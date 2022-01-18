package plugins

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"
	"time"

	"git.icyphox.sh/paprika/database"
	"github.com/dgraph-io/badger/v3"
	"github.com/dustin/go-humanize"
	"gopkg.in/irc.v3"
)

func init() {
	Register(Seen{})
}

type Seen struct{}

func (Seen) Triggers() []string {
	return []string{".seen", ""}
}

type LastSeenInfo struct {
	// The last message the user sent.
	Message string
	// Command type
	Doing string
	// The last time this user was seen.
	Time time.Time
}

func (Seen) Execute(m *irc.Message) (string, error) {
	var reply string
	if m.Command == "PRIVMSG" && strings.HasPrefix(m.Trailing(), ".seen") {
		params := strings.SplitN(m.Trailing(), " ", 3)
		if len(params) != 2 {
			reply = "Usage: .seen <nickname>"
		} else {
			nameKey := database.ToKey("seen", params[1])
			lastS, err := database.DB.Get(nameKey)
			if err == badger.ErrKeyNotFound {
				reply = fmt.Sprintf("I have not seen %s", params[1])
			} else if err != nil {
				return "", err
			} else {
				var lastSeen LastSeenInfo
				err = gob.NewDecoder(bytes.NewReader(lastS)).Decode(&lastSeen)
				if err != nil {
					return "", err
				}

				humanized := humanize.Time(lastSeen.Time)

				if lastSeen.Doing == "PRIVMSG" {
					reply = fmt.Sprintf(
						"\x02%s\x02 was last seen %s, saying: %s",
						params[1], humanized,
						lastSeen.Message,
					)
				} else {
					reply = fmt.Sprintf(
						"\x02%s\x02 was last seen %s, doing: %s",
						params[1], humanized,
						lastSeen.Doing,
					)
				}
			}
		}
	}

	seenDoing := LastSeenInfo{
		Message: m.Trailing(),
		Doing:   m.Command,
		// We just saw the user, so.
		Time: time.Now(),
	}

	var enc bytes.Buffer
	err := gob.NewEncoder(&enc).Encode(seenDoing)
	if err != nil {
		return "", err
	}

	nameKey := database.ToKey("seen", m.Name)
	database.DB.Set(nameKey, enc.Bytes())

	if reply == "" {
		return "", NoReply
	} else {
		return reply, nil
	}
}

func SeenDoing(m *irc.Message) error {
	_, err := Seen{}.Execute(m)
	return err
}
