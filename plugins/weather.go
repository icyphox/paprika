package plugins

import (
	"log"
	"strings"

	"git.icyphox.sh/paprika/plugins/location"
	"git.icyphox.sh/paprika/plugins/time"
	"git.icyphox.sh/paprika/plugins/weather"
	"github.com/dgraph-io/badger/v3"
	"gopkg.in/irc.v3"
)

func init() {
	Register(Weather{})
}

type Weather struct{}

func (Weather) Triggers() []string {
	return []string{
		".w",
		".weather",
		".t",
		".time",
	}
}

func (Weather) Execute(cmd, rest string, c *irc.Client, m *irc.Message) {
	var loc string
	if rest == "" {
		var err error
		// Check if they've already set their location
		loc, err = location.GetLocation(m.Prefix.Name)
		if err == badger.ErrKeyNotFound {
			c.WriteMessage(NewRes(m, "Location not set. Use '.loc <location>' to set it."))
		} else if err != nil {
			log.Panicln(err)
			return
		}
	}

	// They're either querying for a location or @nick.
	if len(loc) == 0 {
		if strings.HasPrefix(rest, "@") {
			// Strip '@'
			var err error
			loc, err = location.GetLocation(rest[1:])
			if err == badger.ErrKeyNotFound {
				c.WriteMessage(NewRes(m, "Location not set. Use '.loc <location>' to set it."))
				return
			} else if err != nil {
				log.Println(err)
				return
			}
		} else {
			loc = rest
		}
	}

	li, err := location.GetLocationInfo(loc)
	if err != nil {
		log.Println(err)
		return
	}

	if len(li.Features) == 0 {
		c.WriteMessage(NewRes(m, "Error getting location info. Try again."))
		return
	}
	coordinates := li.Features[0].Geometry.Coordinates
	label := li.Features[0].Properties.Geocoding.Label

	switch cmd {
	case ".t", ".time":
		time, err := time.GetTime(coordinates, label)
		if err != nil {
			log.Println(err)
			return
		} else {
			c.WriteMessage(NewRes(m, time))
		}
		return
	case ".w", ".weather":

		info, err := weather.GetWeather(coordinates, label)
		if err != nil {
			log.Println(err)
			return
		} else {
			c.WriteMessage(NewRes(m, info))
		}
		return
	}
}
