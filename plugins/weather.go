package plugins

import (
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

func (Weather) Execute(cmd, rest string, m *irc.Message) (*irc.Message, error) {
	var loc string
	if rest == "" {
		var err error
		// Check if they've already set their location
		loc, err = location.GetLocation(m.Prefix.Name)
		if err == badger.ErrKeyNotFound {
			return NewRes(m, "Location not set. Use '.loc <location>' to set it."), nil
		} else if err != nil {
			return nil, err
		}
	}

	// They're either querying for a location or @nick.
	if len(loc) == 0 {
		if strings.HasPrefix(rest, "@") {
			// Strip '@'
			var err error
			loc, err = location.GetLocation(rest[1:])
			if err == badger.ErrKeyNotFound {
				return NewRes(m, "Location not set. Use '.loc <location>' to set it."), nil
			} else if err != nil {
				return nil, err
			}
		} else {
			loc = rest
		}
	}

	li, err := location.GetLocationInfo(loc)
	if err != nil {
		return nil, err
	}

	if len(li.Features) == 0 {
		return NewRes(m, "Error getting location info. Try again."), nil
	}
	coordinates := li.Features[0].Geometry.Coordinates
	label := li.Features[0].Properties.Geocoding.Label

	switch cmd {
	case ".t", ".time":
		time, err := time.GetTime(coordinates, label)
		if err != nil {
			return nil, err
		}
		return NewRes(m, time), nil
	case ".w", ".weather":

		info, err := weather.GetWeather(coordinates, label)
		if err != nil {
			return nil, err
		}
		return NewRes(m, info), nil
	}
	return nil, NoReply
}
