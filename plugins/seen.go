package plugins

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"git.icyphox.sh/paprika/database"
	"github.com/dgraph-io/badger/v3"
	"github.com/dustin/go-humanize"
	"gopkg.in/irc.v3"
)

func init() {
	Register(Seen{})
	RegisterMatcher(Seen{})
}

type Seen struct{}

func (Seen) Triggers() []string {
	return []string{".seen"}
}

type LastSeenInfo struct {
	// The last message the user sent.
	Message string
	// Command type
	Doing string
	// The last time this user was seen.
	Time time.Time
}

func (Seen) Matches(c *irc.Client, m *irc.Message) (string, bool) {
	// always match
	return "", true
}

func (Seen) Execute(cmd, rest string, c *irc.Client, m *irc.Message) {
	var reply string
	if m.Command == "PRIVMSG" && cmd == ".seen" {
		if rest == "" {
			reply = "Usage: .seen <nickname>"
		} else {
			nameKey := database.ToKey("seen", rest)
			lastS, err := database.DB.Get(nameKey)
			if err == badger.ErrKeyNotFound {
				reply = fmt.Sprintf("I have not seen %s", rest)
			} else if err != nil {
				log.Panicln(err)
				return
			} else {
				var lastSeen LastSeenInfo
				err = gob.NewDecoder(bytes.NewReader(lastS)).Decode(&lastSeen)
				if err != nil {
					log.Println(err)
					return
				}

				humanized := humanize.Time(lastSeen.Time)

				if lastSeen.Doing == "PRIVMSG" {
					reply = fmt.Sprintf(
						"\x02%s\x02 was last seen %s, saying: %s",
						rest, humanized,
						lastSeen.Message,
					)
				} else {
					reply = fmt.Sprintf(
						"\x02%s\x02 was last seen %s, doing: %s",
						rest, humanized,
						lastSeen.Doing,
					)
				}
			}
		}
	}

	seenDoing := LastSeenInfo{
		Message: m.Trailing(),
		Doing:   m.Command,
		Time:    time.Now(),
	}

	var enc bytes.Buffer
	err := gob.NewEncoder(&enc).Encode(seenDoing)
	if err != nil {
		log.Println(err)
		return
	}

	nameKey := database.ToKey("seen", m.Name)
	database.DB.Set(nameKey, enc.Bytes())

	if reply != "" {
		c.WriteMessage(NewRes(m, reply))
	}
}

func SeenDoing(c *irc.Client, m *irc.Message) {
	Seen{}.Execute("", "", c, m)
}
