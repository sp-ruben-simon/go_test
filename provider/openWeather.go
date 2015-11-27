package provider

import (
	"encoding/json"
	"log"
	"net/http"
)

type openWeatherData struct {
	Name string `json:"name"`
	Main struct {
		Kelvin float64 `json:"temp"`
	} `json:"main"`
}

type OpenWeatherMap struct {
	ApiKey string
}

func (w OpenWeatherMap) Temperature(city string) (float64, error) {
	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?q=" + city + "&appid=" + w.ApiKey)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	var d openWeatherData

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
	}

	log.Printf("openWeatherMap: %s: %.2f", city, d.Main.Kelvin)
	return d.Main.Kelvin, nil
}
