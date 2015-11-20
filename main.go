package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type multiWeatherProvider []weatherProvider

type weatherProvider interface {
	temperature(city string) (float64, error) //In Kelvin
}

//OpenWeatherMap
type openWeatherData struct {
	Name string `json:"name"`
	Main struct {
		Kelvin float64 `json:"temp"`
	} `json:"main"`
}

type openWeatherMap struct {
	apiKey string
}

func (w openWeatherMap) temperature(city string) (float64, error) {
	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?q=" + city + "&appid=" + w.apiKey)
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

//WeatherUnderground
type weatherUnderground struct {
	apiKey string
}

type weatherUndergroundData struct {
	Observation struct {
		Celsius float64 `json:"temp_c"`
	} `json:"current_observation"`
}

func (w weatherUnderground) temperature(city string) (float64, error) {
	resp, err := http.Get("http://api.wunderground.com/api/" + w.apiKey + "/conditions/q/" + city + ".json")
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

func (w multiWeatherProvider) temperature(city string) (float64, error) {
	temps := make(chan float64, len(w))
	errs := make(chan error, len(w))

	sum := 0.0

	for _, provider := range w {
		go func(p weatherProvider) {
			k, err := provider.temperature(city)
			if err != nil {
				errs <- err
				return
			}
			temps <- k
		}(provider)
	}

	for i := 0; i < len(w); i++ {
		select {
		case temp := <-temps:
			sum += temp
		case err := <-errs:
			return 0, err
		}
	}

	return sum / float64(len(w)), nil
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello!"))
}

func weather(w http.ResponseWriter, r *http.Request) {
	mw := multiWeatherProvider{
		openWeatherMap{apiKey: "2de143494c0b295cca9337e1e96b00e0"},
		weatherUnderground{apiKey: "fa05f5ad8312f4f0"},
	}

	begin := time.Now()
	city := strings.SplitN(r.URL.Path, "/", 3)[2]

	temp, err := mw.temperature(city)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"city": city,
		"temp": temp,
		"took": time.Since(begin).String(),
	})
}

func main() {
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/weather/", weather)

	http.ListenAndServe(":8080", nil)
}
