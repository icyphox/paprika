// User profile plugins: desktop, homescreen, selfie, battlestation, etc.
package plugins

import (
	"fmt"
	"strings"

	"git.icyphox.sh/paprika/database"
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

func (Profile) Execute(m *irc.Message) (string, error) {
	parts := strings.SplitN(m.Trailing(), " ", 2)
	trigger := parts[0]

	var key string

	switch trigger {
	case ".desktop", ".dt":
		key = "desktop"
	case ".hs", ".homescreen", ".home":
		key = "homescreen"
	case ".bs", ".battlestation", ".keyb":
		key = "battlestation"
	default:
		// Strip the '.'
		key = trigger[1:]
	}

	if len(parts) == 1 {
		val, err := database.DB.Get([]byte(fmt.Sprintf(
			"%s/%s",
			key,
			strings.ToLower(m.Prefix.Name),
		)))
		if err != nil {
			return fmt.Sprintf(
				"Error fetching %s. Use '%s <link>' to set it.", key, trigger,
			), err
		}
		return fmt.Sprintf("\x02%s\x02: %s", m.Prefix.Name, string(val)), nil
	} else if len(parts) == 2 {
		// Querying @nick's thing.
		if strings.HasPrefix(parts[1], "@") {
			val, err := database.DB.Get([]byte(fmt.Sprintf(
				"%s/%s",
				key,
				parts[1][1:],
			)))
			if err != nil {
				return fmt.Sprintf("Error fetching %s", key), err
			}
			return fmt.Sprintf("\x02%s\x02: %s", parts[1][1:], string(val)), nil
		}
		// User wants to set the thing.
		value := parts[1]

		err := database.DB.Set(
			[]byte(fmt.Sprintf("%s/%s", key, strings.ToLower(m.Prefix.Name))),
			[]byte(value),
		)
		if err != nil {
			return fmt.Sprintf("Error saving %s", key), err
		}
		return fmt.Sprintf("Saved your %s successfully", key), nil
	}
	return "", nil
}
