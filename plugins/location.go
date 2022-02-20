package plugins

import (
	"fmt"
	"log"

	"git.icyphox.sh/paprika/plugins/location"
	"gopkg.in/irc.v3"
)

func init() {
	Register(Location{})
}

type Location struct{}

func (Location) Triggers() []string {
	return []string{".loc", ".location"}
}

func (Location) Execute(cmd, loc string, c *irc.Client, m *irc.Message) {
	if loc == "" {
		c.WriteMessage(NewRes(m, fmt.Sprintf("Usage: %s <location>", cmd)))
		return
	}

	err := location.SetLocation(loc, m.Prefix.Name)
	if err != nil {
		log.Println(err)
	} else {
		c.WriteMessage(NewRes(m, "Successfully set location"))
	}
}
