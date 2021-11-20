package plugins

import (
	"strings"

	"git.icyphox.sh/paprika/plugins/location"
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
	}
}

func (Weather) Execute(m *irc.Message) (string, error) {
	parsed := strings.SplitN(m.Trailing(), " ", 2)
	var loc string
	if len(parsed) != 2 {
		var err error
		// Check if they've already set their location
		loc, err = location.GetLocation(m.Prefix.Name)
		if err == badger.ErrKeyNotFound {
			return "Location not set. Use '.loc <location>' to set it.", nil
		} else if err != nil {
			return "Error getting location", err
		}
	}

	// They're either querying for a location or @nick.
	if len(loc) == 0 {
		if strings.HasPrefix(parsed[1], "@") {
			// Strip '@'
			var err error
			loc, err = location.GetLocation(parsed[1][1:])
			if err == badger.ErrKeyNotFound {
				return "Location not set. Use '.loc <location>' to set it.", nil
			} else if err != nil {
				return "Error getting location. Try again.", err
			}
		} else {
			loc = parsed[1]
		}
	}

	li, err := location.GetLocationInfo(loc)
	if err != nil {
		return "Error getting location info. Try again.", err
	}

	if len(li.Features) == 0 {
		return "Error getting location info. Try again.", nil
	}
	coordinates := li.Features[0].Geometry.Coordinates
	label := li.Features[0].Properties.Geocoding.Label
	info, err := weather.GetWeather(coordinates, label)
	if err != nil {
		return "Error getting weather data", err
	}
	return info, nil
}
