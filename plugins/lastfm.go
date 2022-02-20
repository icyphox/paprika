package plugins

import (
	"log"

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

func (LastFM) Execute(cmd, rest string, c *irc.Client, m *irc.Message) {
	switch cmd {
	case ".lfm":
		if rest != "" {
			err := lastfm.Setup(rest, m.Prefix.Name)
			if err != nil {
				log.Println(err)
				return
			}
			c.WriteMessage(NewRes(m, "Successfully set Last.fm username"))
		} else {
			c.WriteMessage(NewRes(m, "Usage: .lfm <username>"))
		}
	case ".np":
		user, err := lastfm.GetUser(m.Prefix.Name)
		if err == badger.ErrKeyNotFound {
			c.WriteMessage(NewRes(m, "User not found. Set it using '.lfm <username>'"))
			return
		} else if err != nil {
			log.Println(err)
			return
		}

		np, err := lastfm.NowPlaying(user)
		if err != nil {
			log.Println(err)
			return
		}
		c.WriteMessage(NewRes(m, np))
	}
}
