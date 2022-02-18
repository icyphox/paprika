package plugins

import (
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

func (LastFM) Execute(cmd, rest string, m *irc.Message) (*irc.Message, error) {
	switch cmd {
	case ".lfm":
		if rest != "" {
			err := lastfm.Setup(rest, m.Prefix.Name)
			if err != nil {
				return nil, err
			}
			return NewRes(m, "Successfully set Last.fm username"), nil
		} else {
			return NewRes(m, "Usage: .lfm <username>"), nil
		}
	case ".np":
		user, err := lastfm.GetUser(m.Prefix.Name)
		if err == badger.ErrKeyNotFound {
			return NewRes(m, "User not found. Set it using '.lfm <username>'"), nil
		} else if err != nil {
			return nil, err
		}

		np, err := lastfm.NowPlaying(user)
		if err != nil {
			return nil, err
		}
		return NewRes(m, np), nil
	}

	panic("Unreachable!")
}
