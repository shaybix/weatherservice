package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type TemperatureAPI struct {
	URL     string
	Date    string
	wg      *sync.WaitGroup
	resChan chan TemperatureResponse
}

// TemperatureResponse holds the temperature of a given date.
type TemperatureResponse struct {
	Temp  float64 `json:"temp,omitempty"`
	Date  string  `json:"date,omitempty"`
	Error string  `json:"error,omitempty"`
}

// TemperatureHandler is the handler for the /temperatures endpoint.
func (tapi TemperatureAPI) TemperatureHandler(w http.ResponseWriter, r *http.Request) {

	// Get the start and end variables from the query string,
	// and handle error if not present.
	var (
		start string
		end   string
	)

	// create a buffered channel that receives Temperature structs
	tapi.resChan = make(chan TemperatureResponse, 5)

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
	tapi.wg = &wg

	// daysBetween := endDate.Sub(startDate).Hours() / 24
	jobsChan := make(chan APIGetter, 3)
	wp := WorkerPool{jobsChan: jobsChan, rateLimit: 30}

Loop:
	for d := startDate; ; d = d.AddDate(0, 0, 1) {
		if d.Year() == endDate.Year() && d.Month() == endDate.Month() && d.Day() == endDate.Day() {
			wg.Add(1)
			tapi.Date = d.Format(time.RFC3339)
			worker := Worker{apiRequest: tapi}
			wp.queue = append(wp.queue, worker)

			break Loop
		}
		wg.Add(1)
		tapi.Date = d.Format(time.RFC3339)
		worker := Worker{apiRequest: tapi}
		wp.queue = append(wp.queue, worker)
	}

	go wp.Run()

	var temp []TemperatureResponse
Loop2:
	for {

		select {
		case w := <-wp.jobsChan:
			go w.Get()
			time.Sleep(time.Second / wp.rateLimit)
		case t := <-tapi.resChan:
			if t.Date == "" {
				t.Error = fmt.Sprint("Data not available for given date")
			}
			temp = append(temp, t)
		case <-time.After(time.Millisecond * 500):
			break Loop2
		}
	}

	if err := json.NewEncoder(w).Encode(temp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get ...
func (tapi TemperatureAPI) Get() {

	var temp TemperatureResponse
	defer tapi.wg.Done()

	resp, err := http.Get(fmt.Sprintf("%s%s", tapi.URL, tapi.Date))
	if err != nil {
		temp.Error = err.Error()
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&temp); err != nil {
		temp.Error = err.Error()
	}

	tapi.resChan <- temp
	return
}
