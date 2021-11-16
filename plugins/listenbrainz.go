// Listenbrainz now-playing info.
package plugins

import (
	"strings"

	"git.icyphox.sh/taigobot/plugins/listenbrainz"
	"github.com/dgraph-io/badger/v3"
	"gopkg.in/irc.v3"
)

func init() {
	Register(Listenbrainz{})
}

type Listenbrainz struct{}

func (Listenbrainz) Triggers() []string {
	return []string{".np", ".lbz"}
}

func (Listenbrainz) Execute(m *irc.Message) (string, error) {
	parts := strings.SplitN(m.Trailing(), " ", 2)
	trigger := parts[0]

	switch trigger {
	case ".lbz":
		if len(parts) == 2 {
			arg := parts[1]
			err := listenbrainz.Setup(arg, m.Prefix.Name)
			if err != nil {
				return "Database error", err
			}
			return "Successfully set Listenbrainz username", nil
		} else {
			return "Usage: .lbz <username>", nil
		}
	case ".np":
		user, err := listenbrainz.GetUser(m.Prefix.Name)
		if err == badger.ErrKeyNotFound {
			return "User not found. Set it using '.lbz <username>'", nil
		} else if err != nil {
			return "Database error", err
		}

		np, err := listenbrainz.NowPlaying(user)
		if err != nil {
			return "Listenbrainz error", err
		}

		return np, nil
	}

	return "", nil
}
