package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type LocationInfo struct {
	Features []struct {
		Properties struct {
			Geocoding struct {
				Label string `json:"label"`
			} `json:"geocoding"`
		} `json:"properties"`
		Geometry struct {
			Coordinates [2]float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}

// Get location data (lat,lon) from a given query.
func GetLocationInfo(query string) (*LocationInfo, error) {
	url := "https://nominatim.openstreetmap.org/search?q=%s&format=geocodejson&limit=1"
	r, err := http.Get(fmt.Sprintf(url, query))
	if err != nil {
		return nil, err
	}

	li := LocationInfo{}
	json.NewDecoder(r.Body).Decode(&li)
	defer r.Body.Close()

	return &li, nil
}
