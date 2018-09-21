package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// WeatherAPI ...
type WeatherAPI struct {
	TemperatureURL string
	WindSpeedURL   string
	Date           string
	wg             *sync.WaitGroup
	resChan        chan WeatherResponse
}

// WeatherResponse holds the temperature of a given date.
type WeatherResponse struct {
	North float64 `json:"north,omitempty"`
	West  float64 `json:"west,omitempty"`
	Temp  float64 `json:"temp,omitempty"`
	Date  string  `json:"date,omitempty"`
	Error string  `json:"error,omitempty"`
}

// WeatherHandler ...
func (wapi WeatherAPI) WeatherHandler(w http.ResponseWriter, r *http.Request) {

	// Get the start and end variables from the query string,
	// and handle error if not present.
	var (
		start string
		end   string
	)

	// create a buffered channel that receives Weather structs
	wapi.resChan = make(chan WeatherResponse, 5)

	// Exctracts the query params from the url.
	values := r.URL.Query()

	// Get the start and end query params and assign them to variables.
	start = values.Get("start")
	end = values.Get("end")

	// iterate over the date range and execute the requests in  a goroutine.
	// TODO: iterate over date range given.
	startDate, err := ISO8601ToTime(start)
	if err != nil {
		er := Error{Error: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(er)
		return
	}
	endDate, err := ISO8601ToTime(end)
	if err != nil {
		er := Error{Error: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(er)
		return
	}
	var wg sync.WaitGroup
	wapi.wg = &wg

	// daysBetween := endDate.Sub(startDate).Hours() / 24
	jobsChan := make(chan APIGetter, 3)
	wp := WorkerPool{jobsChan: jobsChan, rateLimit: 30}

Loop:
	for d := startDate; ; d = d.AddDate(0, 0, 1) {
		if d.Year() == endDate.Year() && d.Month() == endDate.Month() && d.Day() == endDate.Day() {
			wg.Add(1)
			wapi.Date = d.Format(time.RFC3339)
			worker := Worker{apiRequest: wapi}
			wp.queue = append(wp.queue, worker)
			// go tapi.Get()

			break Loop
		}
		wg.Add(1)
		wapi.Date = d.Format(time.RFC3339)
		worker := Worker{apiRequest: wapi}
		wp.queue = append(wp.queue, worker)
		// go tapi.Get()
	}

	go wp.Run()

	var weathers []WeatherResponse
Loop2:
	for {

		select {
		case w := <-wp.jobsChan:
			go w.Get()
			time.Sleep(time.Second / wp.rateLimit)
		case t := <-wapi.resChan:
			if t.Date == "" {
				t.Error = fmt.Sprint("Data not available for given date")
			}
			weathers = append(weathers, t)
		case <-time.After(time.Millisecond * 500):
			break Loop2
		}
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(weathers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

// Get ...
func (wapi WeatherAPI) Get() {

	var (
		tempResp TemperatureResponse
		windResp WindSpeedResponse
	)

	defer wapi.wg.Done()

	tresp, err := http.Get(fmt.Sprintf("%s%s", wapi.TemperatureURL, wapi.Date))
	if err != nil {
		log.Println(err)
		return
	}
	defer tresp.Body.Close()
	if err := json.NewDecoder(tresp.Body).Decode(&tempResp); err != nil {
		log.Println(err)
		return
	}

	wndresp, err := http.Get(fmt.Sprintf("%s%s", wapi.WindSpeedURL, wapi.Date))
	if err != nil {
		log.Println(err)
		return
	}
	defer wndresp.Body.Close()
	if err := json.NewDecoder(wndresp.Body).Decode(&windResp); err != nil {
		log.Println(err)
		return
	}

	wresp := WeatherResponse{
		North: windResp.North,
		West:  windResp.West,
		Temp:  tempResp.Temp,
		Date:  wapi.Date,
	}

	wapi.resChan <- wresp
	return
}
