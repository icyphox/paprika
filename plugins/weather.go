package plugins

import (
	"fmt"
	"strings"

	"git.icyphox.sh/paprika/plugins/weather"
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
	if len(parsed) != 2 {
		return fmt.Sprintf("Usage: %s <location>", parsed[0]), nil
	}
	query := parsed[1]
	li, err := weather.GetLocationInfo(query)
	if err != nil {
		return "Error getting location info", err
	}

	coordinates := li.Features[0].Geometry.Coordinates
	label := li.Features[0].Properties.Geocoding.Label
	info, err := weather.GetWeather(coordinates, label)
	if err != nil {
		return "Error getting weather data", err
	}
	return info, nil
}
