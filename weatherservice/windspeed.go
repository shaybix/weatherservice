package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// WindSpeedResponse ...
type WindSpeedResponse struct {
	North float64 `json:"north,omitempty"`
	West  float64 `json:"west,omitempty"`
	Date  string  `json:"date,omitempty"`
	Error string  `json:"error,omitempty"`
}

// WindSpeedAPI ...
type WindSpeedAPI struct {
	URL     string
	Date    string
	wg      *sync.WaitGroup
	resChan chan WindSpeedResponse
}

// WindSpeedHandler ...
func (wsapi WindSpeedAPI) WindSpeedHandler(w http.ResponseWriter, r *http.Request) {

	// Get the start and end variables from the query string,
	// and handle error if not present.
	var (
		start string
		end   string
	)

	// create a buffered channel that receives Temperature structs
	wsapi.resChan = make(chan WindSpeedResponse, 5)

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
	wsapi.wg = &wg

	// daysBetween := endDate.Sub(startDate).Hours() / 24
	jobsChan := make(chan APIGetter, 3)
	wp := WorkerPool{jobsChan: jobsChan, rateLimit: 30}

Loop:
	for d := startDate; ; d = d.AddDate(0, 0, 1) {
		if d.Year() == endDate.Year() && d.Month() == endDate.Month() && d.Day() == endDate.Day() {
			wg.Add(1)
			wsapi.Date = d.Format(time.RFC3339)
			worker := Worker{apiRequest: wsapi}
			wp.queue = append(wp.queue, worker)
			// go tapi.Get()

			break Loop
		}
		wg.Add(1)
		wsapi.Date = d.Format(time.RFC3339)
		worker := Worker{apiRequest: wsapi}
		wp.queue = append(wp.queue, worker)
		// go tapi.Get()
	}

	go wp.Run()

	var wspeeds []WindSpeedResponse
Loop2:
	for {

		select {
		case w := <-wp.jobsChan:
			go w.Get()
			time.Sleep(time.Second / wp.rateLimit)
		case t := <-wsapi.resChan:
			if t.Date == "" {
				t.Error = fmt.Sprint("Data not available for given date")
			}
			wspeeds = append(wspeeds, t)
		case <-time.After(time.Millisecond * 500):
			break Loop2
		}
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(wspeeds); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

// Get ...
func (wsapi WindSpeedAPI) Get() {

	var wsresp WindSpeedResponse
	defer wsapi.wg.Done()

	resp, err := http.Get(fmt.Sprintf("%s%s", wsapi.URL, wsapi.Date))
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&wsresp); err != nil {
		log.Println(err)
		return
	}

	wsapi.resChan <- wsresp
	return
}
