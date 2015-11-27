package provider

import (
	"encoding/json"
	"log"
	"net/http"
)

type WeatherUnderground struct {
	ApiKey string
}

type weatherUndergroundData struct {
	Observation struct {
		Celsius float64 `json:"temp_c"`
	} `json:"current_observation"`
}

func (w WeatherUnderground) Temperature(city string) (float64, error) {
	resp, err := http.Get("http://api.wunderground.com/api/" + w.ApiKey + "/conditions/q/" + city + ".json")
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	var d weatherUndergroundData

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
	}

	kelvin := d.Observation.Celsius + 273.15
	log.Printf("weatherUnderground: %s: %.2f", city, kelvin)
	return kelvin, nil
}
