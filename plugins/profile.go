// User profile plugins: desktop, homescreen, selfie, battlestation, etc.
package plugins

import (
	"fmt"
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

func (Profile) Execute(cmd, rest string, m *irc.Message) (*irc.Message, error) {
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
			return NewRes(m, fmt.Sprintf(
				"Error fetching %s. Use '%s <link>' to set it.", key, cmd,
			)), nil
		} else if err != nil {
			return nil, err
		}

		return NewRes(m, fmt.Sprintf("\x02%s\x02: %s", m.Prefix.Name, string(val))), nil
	} else {
		// Querying @nick's thing.
		if rest[0] == '@' {
			val, err := database.DB.Get(database.ToKey(key, rest[1:]))
			if err == badger.ErrKeyNotFound {
				return NewRes(m, fmt.Sprintf("No %s for %s", key, rest[1:])), err
			} else if err != nil {
				return nil, err
			}
			return NewRes(m, fmt.Sprintf("\x02%s\x02: %s", rest[1:], string(val))), nil
		}
		// User wants to set the thing.

		err := database.DB.Set(
			database.ToKey(key, strings.ToLower(m.Prefix.Name)),
			[]byte(rest),
		)
		if err != nil {
			return nil, err
		}
		return NewRes(m, fmt.Sprintf("Saved your %s successfully", key)), nil
	}
}
