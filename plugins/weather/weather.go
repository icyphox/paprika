package weather

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type WeatherData struct {
	Properties struct {
		Meta struct {
			Units struct {
				AirPressureAtSeaLevel string `json:"air_pressure_at_sea_level"`
				AirTemperature        string `json:"air_temperature"`
				RelativeHumidity      string `json:"relative_humidity"`
				WindSpeed             string `json:"wind_speed"`
			} `json:"units"`
		} `json:"meta"`
		Timeseries []struct {
			Data struct {
				Instant struct {
					Details struct {
						AirPressureAtSeaLevel float64 `json:"air_pressure_at_sea_level"`
						AirTemperature        float64 `json:"air_temperature"`
						RelativeHumidity      float64 `json:"relative_humidity"`
						WindSpeed             float64 `json:"wind_speed"`
					} `json:"details"`
				} `json:"instant"`
				Next12Hours struct {
					Summary struct {
						SymbolCode string `json:"symbol_code"`
					} `json:"summary"`
				} `json:"next_12_hours"`
			} `json:"data"`
		} `json:"timeseries"`
	} `json:"properties"`
}

func getWeatherData(lonlat [2]float64) (*WeatherData, error) {
	url := "https://api.met.no/weatherapi/locationforecast/2.0/compact.json?lat=%f16&lon=%f"
	url = fmt.Sprintf(url, lonlat[1], lonlat[0])

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "taigobot github.com/icyphox/taigobot")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	wd := WeatherData{}

	json.NewDecoder(res.Body).Decode(&wd)
	defer res.Body.Close()

	return &wd, nil
}

func ctof(c float64) float64 {
	return c*9/5 + 32
}

//go:embed data/weather.json
var wj []byte

func symbolToDesc(s string) string {
	m := make(map[string]string)
	json.Unmarshal(wj, &m)

	// Some symbols have a _, which isn't present in
	// the data. So we strip it.
	s = strings.Split(s, "_")[0]

	return m[s] + ", "
}

// Looks like Nominatim uses (lon,lat).
func GetWeather(lonlat [2]float64, location string) (string, error) {
	wd, err := getWeatherData(lonlat)
	if err != nil {
		return "", err
	}

	info := strings.Builder{}

	// Location.
	fmt.Fprintf(&info, "\x02%s\x02: ", location)

	// Description of weather.
	sym := wd.Properties.Timeseries[0].Data.Next12Hours.Summary.SymbolCode
	desc := symbolToDesc(sym)
	fmt.Fprintf(&info, "%s", desc)

	// Current temperature.
	temp := wd.Properties.Timeseries[0].Data.Instant.Details.AirTemperature
	var tempFmt string
	if temp > 28 {
		tempFmt = "\x0304%0.2f°C\x03 (\x0304%0.2f°F\x03), "
	} else if temp > 18 {
		tempFmt = "\x0303%0.2f°C\x03 (\x0303%0.2f°F\x03), "
	} else {
		tempFmt = "\x0311%0.2f°C\x03 (\x0311%0.2f°F\x03), "
	}
	fmt.Fprintf(
		&info,
		"\x02Currently:\x02 "+tempFmt,
		temp,
		ctof(temp),
	)

	// Humidity.
	hum := wd.Properties.Timeseries[0].Data.Instant.Details.RelativeHumidity
	humUnit := wd.Properties.Meta.Units.RelativeHumidity
	fmt.Fprintf(
		&info,
		"\x02Humidity:\x02 %0.2f%s, ",
		hum,
		humUnit,
	)

	// Wind speed.
	ws := wd.Properties.Timeseries[0].Data.Instant.Details.WindSpeed
	wsUnit := wd.Properties.Meta.Units.WindSpeed
	fmt.Fprintf(
		&info,
		"\x02Wind Speed:\x02 %0.1f %s, ",
		ws,
		wsUnit,
	)

	// Pressure.
	ps := wd.Properties.Timeseries[0].Data.Instant.Details.AirPressureAtSeaLevel
	psUnit := wd.Properties.Meta.Units.AirPressureAtSeaLevel
	fmt.Fprintf(&info,
		"\x02Pressure:\x02 %0.1f %s",
		ps,
		psUnit,
	)

	return info.String(), nil
}
