// Listenbrainz now-playing info.
package plugins

import (
	"git.icyphox.sh/paprika/plugins/listenbrainz"
	"github.com/dgraph-io/badger/v3"
	"gopkg.in/irc.v3"
)

func init() {
	Register(Listenbrainz{})
}

type Listenbrainz struct{}

func (Listenbrainz) Triggers() []string {
	// TODO: removing .np from here until we figure out how
	// it can co-exist with Last.fm
	return []string{".lbz"}
}

func (Listenbrainz) Execute(cmd, msg string, m *irc.Message) (*irc.Message, error) {
	switch cmd {
	case ".lbz":
		if msg == "" {
			err := listenbrainz.Setup(msg, m.Prefix.Name)
			if err != nil {
				return nil, err
			}
			return NewRes(m, "Successfully set Listenbrainz username"), nil
		} else {
			return NewRes(m, "Usage: .lbz <username>"), nil
		}
	case ".np":
		user, err := listenbrainz.GetUser(m.Prefix.Name)
		if err == badger.ErrKeyNotFound {
			return NewRes(m, "User not found. Set it using '.lbz <username>'"), nil
		} else if err != nil {
			return nil, err
		}

		np, err := listenbrainz.NowPlaying(user)
		if err != nil {
			return nil, err
		}

		return NewRes(m, np), nil
	}

	panic("Unreachable!")
}
