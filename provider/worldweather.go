package provider

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Data struct {
	Condition []Condition `json:"current_condition"`
}

type Condition struct {
	Celsius    string `json:"temp_C"`
	Fahrenheit string `json:"temp_F"`
}

type worldWeatherData struct {
	Data Data `json:"data"`
}

type WorldWeather struct {
	ApiKey string
}

func (w WorldWeather) Temperature(country string, city string) (float64, error) {
	resp, err := http.Get("http://api.worldweatheronline.com/free/v2/weather.ashx?q=" + city + "," + country + "&key=" + w.ApiKey + "&format=json")
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	var d worldWeatherData

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
	}

	strCelsius := d.Data.Condition[0].Celsius
	celsius, err := strconv.ParseFloat(strCelsius, 64)
	kelvin := celsius + 273.15
	log.Printf("worldWeather: %s - %s: %.2fC - %.2fK", country, city, celsius, kelvin)
	return celsius, nil
}
