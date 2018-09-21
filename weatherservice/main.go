package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	// TemperatureURL represents the URL of the temperature service inside docker
	TemperatureURL = "http://temperature:8000?at="

	// WindSpeedURL represents the URL of the windspeed service inside docker
	WindSpeedURL = "http://windspeed:8080?at="
)

func main() {

	router := mux.NewRouter()
	tapi := TemperatureAPI{URL: TemperatureURL}
	wsapi := WindSpeedAPI{URL: WindSpeedURL}
	wapi := WeatherAPI{TemperatureURL: TemperatureURL, WindSpeedURL: WindSpeedURL}

	router.HandleFunc("/temperatures", tapi.TemperatureHandler).Methods("GET")
	router.HandleFunc("/speeds", wsapi.WindSpeedHandler).Methods("GET")
	router.HandleFunc("/weather", wapi.WeatherHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(":8888", router))
	return
}
