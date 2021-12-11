package plugins

import (
	"strings"

	"git.icyphox.sh/paprika/plugins/lastfm"
	"github.com/dgraph-io/badger/v3"
	"gopkg.in/irc.v3"
)

func init() {
	Register(LastFM{})
}

type LastFM struct{}

func (LastFM) Triggers() []string {
	return []string{
		".lfm",
		".np",
	}
}

func (LastFM) Execute(m *irc.Message) (string, error) {
	parts := strings.SplitN(m.Trailing(), " ", 2)
	trigger := parts[0]

	switch trigger {
	case ".lfm":
		if len(parts) == 2 {
			arg := parts[1]
			err := lastfm.Setup(arg, m.Prefix.Name)
			if err != nil {
				return "Database error", err
			}
			return "Successfully set Last.fm username", nil
		} else {
			return "usage: .lfm <username>", nil
		}
	case ".np":
		user, err := lastfm.GetUser(m.Prefix.Name)
		if err == badger.ErrKeyNotFound {
			return "User not found. Set it using '.lfm <username>'", nil
		} else if err != nil {
			return "Database error", err
		}

		np, err := lastfm.NowPlaying(user)
		if err != nil {
			return "Listenbrainz error", err
		}
		return np, nil
	}

	panic("Unreachable!")
}
