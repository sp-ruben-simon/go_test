package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sp-ruben-simon/go_test/provider"
)

type multiWeatherProvider []weatherProvider

type weatherProvider interface {
	Temperature(country string, city string) (float64, error) //In Celsius
}

func (mw multiWeatherProvider) temperature(country string, city string) (float64, error) {
	temps := make(chan float64, len(mw))
	errs := make(chan error, len(mw))

	sum := 0.0

	for _, provider := range mw {
		go func(p weatherProvider) {
			k, err := p.Temperature(country, city)
			if err != nil {
				errs <- err
				return
			}
			temps <- k
		}(provider)
	}

	for i := 0; i < len(mw); i++ {
		select {
		case temp := <-temps:
			sum += temp
		case err := <-errs:
			return 0, err
		}
	}

	return sum / float64(len(mw)), nil
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	mw := multiWeatherProvider{
		provider.OpenWeatherMap{"f78e8c405bea7a313ba80a48046063a8"},
		provider.WeatherUnderground{"fa05f5ad8312f4f0"},
		provider.WorldWeather{"91ae5863c85227a15757aa5bd1343"},
	}

	begin := time.Now()
	vars := mux.Vars(r)
	country := vars["country"]
	city := vars["city"]

	log.Printf("Get weather from %s/%s", country, city)

	temp, err := mw.temperature(country, city)
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
	r := mux.NewRouter()
	r.HandleFunc("/weather/{country}/{city}", weatherHandler)

	http.ListenAndServe(":8080", r)
}
