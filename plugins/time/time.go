package time

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/zsefvlol/timezonemapper"
)

func getTimezone(lonlat [2]float64) string {
	tz := timezonemapper.LatLngToTimezoneString(lonlat[1], lonlat[0])
	return tz
}

func GetTime(lonlat [2]float64, label string) (string, error) {
	url := path.Join(
		"worldtimeapi.org/api/timezone/",
		getTimezone(lonlat),
	)

	m := make(map[string]string)
	r, err := http.Get("https://" + url)
	if err != nil {
		return "", err
	}

	json.NewDecoder(r.Body).Decode(&m)
	timestamp := m["datetime"]
	t, _ := time.Parse(time.RFC3339Nano, timestamp)
	pretty := t.Format("2006-02-01 03:04 PM")
	out := fmt.Sprintf("\x02%s\x02: %s", label, pretty)
	return out, nil
}
