package time

import (
	"fmt"
	"time"

	"github.com/zsefvlol/timezonemapper"
)

func getTimezone(lonlat [2]float64) string {
	tz := timezonemapper.LatLngToTimezoneString(lonlat[1], lonlat[0])
	return tz
}

func GetTime(lonlat [2]float64, label string) (string, error) {
	loc, _ := time.LoadLocation(getTimezone(lonlat))
	now := time.Now().In(loc)
	pretty := now.Format("2006-02-01 03:04 PM")
	out := fmt.Sprintf("\x02%s\x02: %s", label, pretty)
	return out, nil
}
