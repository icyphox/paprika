// User profile plugins: desktop, homescreen, selfie, battlestation, etc.
package plugins

import (
	"fmt"
	"log"
	"strings"

	"git.icyphox.sh/paprika/database"
	"github.com/dgraph-io/badger/v3"
	"gopkg.in/irc.v3"
)

func init() {
	Register(Profile{})
}

type Profile struct{}

func (Profile) Triggers() []string {
	return []string{
		".selfie",
		".desktop", ".dt",
		".hs", ".homescreen", ".home",
		".bs", ".battlestation", ".keyb",
	}
}

func (Profile) Execute(cmd, rest string, c *irc.Client, m *irc.Message) {
	var key string

	switch cmd {
	case ".desktop", ".dt":
		key = "desktop"
	case ".hs", ".homescreen", ".home":
		key = "homescreen"
	case ".bs", ".battlestation", ".keyb":
		key = "battlestation"
	default:
		// Strip the '.'
		key = cmd[1:]
	}

	if rest != "" {
		val, err := database.DB.Get(database.ToKey(key, strings.ToLower(m.Prefix.Name)))
		if err == badger.ErrKeyNotFound {
			c.WriteMessage(NewRes(m, fmt.Sprintf(
				"Error fetching %s. Use '%s <link>' to set it.", key, cmd,
			)))
		} else if err != nil {
			log.Println(err)
			return
		}

		c.WriteMessage(NewRes(m, fmt.Sprintf("\x02%s\x02: %s", m.Prefix.Name, string(val))))
		return
	} else {
		// Querying @nick's thing.
		if rest[0] == '@' {
			val, err := database.DB.Get(database.ToKey(key, rest[1:]))
			if err == badger.ErrKeyNotFound {
				c.WriteMessage(NewRes(m, fmt.Sprintf("No %s for %s", key, rest[1:])))
				return
			} else if err != nil {
				log.Println(err)
				return
			}
			c.WriteMessage(NewRes(m, fmt.Sprintf("\x02%s\x02: %s", rest[1:], string(val))))
			return
		}
		// User wants to set the thing.

		err := database.DB.Set(
			database.ToKey(key, strings.ToLower(m.Prefix.Name)),
			[]byte(rest),
		)
		if err != nil {
			log.Println(err)
		} else {
			c.WriteMessage(NewRes(m, fmt.Sprintf("Saved your %s successfully", key)))
		}
	}
}
