// Listenbrainz now-playing info.
package plugins

import (
	"log"

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

func (Listenbrainz) Execute(cmd, msg string, c *irc.Client, m *irc.Message) {
	switch cmd {
	case ".lbz":
		if msg == "" {
			err := listenbrainz.Setup(msg, m.Prefix.Name)
			if err != nil {
				log.Println(err)
			} else {
				c.WriteMessage(NewRes(m, "Successfully set Listenbrainz username"))
			}
		} else {
			c.WriteMessage(NewRes(m, "Usage: .lbz <username>"))
		}
	case ".np":
		user, err := listenbrainz.GetUser(m.Prefix.Name)
		if err == badger.ErrKeyNotFound {
			c.WriteMessage(NewRes(m, "User not found. Set it using '.lbz <username>'"))
			return
		} else if err != nil {
			log.Println(err)
			return
		}

		np, err := listenbrainz.NowPlaying(user)
		if err != nil {
			log.Println(err)
		} else {
			c.WriteMessage(NewRes(m, np))
		}
	}
}
