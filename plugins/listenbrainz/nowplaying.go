package listenbrainz

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dustin/go-humanize"
)

type ListenInfo struct {
	Payload struct {
		Count   int `json:"count"`
		Listens []struct {
			InsertedAt    string `json:"inserted_at"`
			ListenedAt    int    `json:"listened_at"`
			TrackMetadata struct {
				ArtistName  string `json:"artist_name"`
				ReleaseName string `json:"release_name"`
				TrackName   string `json:"track_name"`
			} `json:"track_metadata"`
			UserName string `json:"user_name"`
		} `json:"listens"`
	} `json:"payload"`
}

func getListen(url string) (*ListenInfo, error) {
	li := ListenInfo{}
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	} else if r.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 response from Listenbrainz")
	}

	json.NewDecoder(r.Body).Decode(&li)
	defer r.Body.Close()

	return &li, err
}

func NowPlaying(user string) (string, error) {
	url := fmt.Sprintf(
		"https://api.listenbrainz.org/1/user/%s/playing-now",
		user,
	)
	li, err := getListen(url)
	if err != nil {
		return "", err
	}

	if len(li.Payload.Listens) != 0 {
		// Now playing a track
		tm := li.Payload.Listens[0].TrackMetadata
		return fmt.Sprintf(
			"%s is currently listening to \"%s\" by \x02%s\x02, from the album \x02%s\x02",
			user,
			tm.TrackName,
			tm.ArtistName,
			tm.ReleaseName,
		), nil
	} else {
		// Last listen
		url = fmt.Sprintf(
			"https://api.listenbrainz.org/1/user/%s/listens?count=1",
			user,
		)
		li, err := getListen(url)
		if err != nil {
			return "", err
		}

		tm := li.Payload.Listens[0].TrackMetadata
		listenedAt, _ := time.Parse(time.RFC1123, li.Payload.Listens[0].InsertedAt)
		return fmt.Sprintf(
			"%s listened to \"%s\" by \x02%s\x02, from the album \x02%s\x02, %s",
			user,
			tm.TrackName,
			tm.ArtistName,
			tm.ReleaseName,
			humanize.Time(listenedAt),
		), nil
	}
}
