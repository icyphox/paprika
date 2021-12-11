package lastfm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"git.icyphox.sh/paprika/config"
	"github.com/dustin/go-humanize"
)

type ListenInfo struct {
	RecentTracks struct {
		Track []struct {
			Artist struct {
				Text string `json:"#text"`
			} `json:"artist"`
			Album struct {
				Text string `json:"#text"`
			} `json:"album"`
			Name string `json:"name"`
			Date struct {
				UnixTimestamp string `json:"uts"`
			} `json:"date"`
			Attr struct {
				NowPlaying string `json:"nowplaying"`
			} `json:"@attr"`
		} `json:"track"`
	} `json:"recenttracks"`
}

func getRecentTracks(url string) (*ListenInfo, error) {
	li := ListenInfo{}
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	} else if r.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 response from Last.fm")
	}

	json.NewDecoder(r.Body).Decode(&li)
	defer r.Body.Close()

	return &li, err
}

func NowPlaying(user string) (string, error) {
	key := config.C.ApiKeys["lastfm-key"]
	url := fmt.Sprintf(
		"https://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks&user=%s&api_key=%s&format=json",
		user,
		key,
	)

	rt, err := getRecentTracks(url)
	if err != nil {
		return "", err
	}

	track := rt.RecentTracks.Track[0]
	if rt.RecentTracks.Track[0].Attr.NowPlaying == "true" {
		return fmt.Sprintf(
			"%s is currently listening to \"%s\" by \x02%s\x02, from the album \x02%s\x02",
			user,
			track.Name,
			track.Artist.Text,
			track.Album.Text,
		), nil
	} else {
		strT := track.Date.UnixTimestamp
		ts, _ := strconv.Atoi(strT)
		t := time.Unix(int64(ts), 0)
		return fmt.Sprintf(
			"%s listened to \"%s\" by \x02%s\x02, from the album \x02%s\x02, %s",
			user,
			track.Name,
			track.Artist.Text,
			track.Album.Text,
			humanize.Time(t),
		), nil
	}
}
