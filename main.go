package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sp-ruben-simon/go_test/provider"
)

type multiWeatherProvider []weatherProvider

type weatherProvider interface {
	Temperature(city string) (float64, error) //In Kelvin
}

func (w multiWeatherProvider) temperature(city string) (float64, error) {
	temps := make(chan float64, len(w))
	errs := make(chan error, len(w))

	sum := 0.0

	for _, provider := range w {
		go func(p weatherProvider) {
			k, err := provider.Temperature(city)
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

func weather(w http.ResponseWriter, r *http.Request) {
	mw := multiWeatherProvider{
		provider.OpenWeatherMap{"2de143494c0b295cca9337e1e96b00e0"},
		provider.WeatherUnderground{"fa05f5ad8312f4f0"},
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
	http.HandleFunc("/weather/", weather)

	http.ListenAndServe(":8080", nil)
}
