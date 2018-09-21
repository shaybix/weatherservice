package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWeatherHandler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tt := []struct {
		name      string
		startDate string
		endDate   string
		status    int
		url       string
	}{
		{"Bad Request", "Monday 16 September 2017", "Wednesday 18 September 2017", http.StatusBadRequest, server.URL},
		{"Bad Request end date", "2018-08-03T00:00:00Z", "Wednesday 18 September 2017", http.StatusBadRequest, server.URL},
		{"End date missing date", "2018-08-03T00:00:00Z", "", http.StatusBadRequest, server.URL},
		{"start date missing date", "", "2018-08-03T00:00:00Z", http.StatusBadRequest, server.URL},
		{"Expecting HTTP OK ", "2018-05-10T00:00:00Z", "2018-05-15T00:00:00Z", http.StatusOK, server.URL},
		{"Bad Request dates in future", "2018-12-10T00:00:00Z", "2018-12-18T00:00:00Z", http.StatusBadRequest, server.URL},
		{"Destination of API endpoints non-existent", "2018-12-10T00:00:00Z", "2018-12-18T00:00:00Z", http.StatusBadRequest, "badhttpaddress"},
	}

	for _, tc := range tt {

		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8888/?start=%s&end=%s", tc.startDate, tc.endDate), nil)
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}
			rec := httptest.NewRecorder()

			weath := WeatherAPI{
				TemperatureURL: fmt.Sprintf("%s?at=%s", server.URL, tc.startDate),
				WindSpeedURL:   fmt.Sprintf("%s?at=%s", server.URL, tc.startDate),
			}
			weath.WeatherHandler(rec, req)

			resp := rec.Result()

			if resp.StatusCode != tc.status {
				t.Errorf("Expected HTTP Status %v but received %v", tc.status, resp.StatusCode)
			}

			// checking that response body contains an error in the response should
			// the response be anything other than the http OK status.
			if resp.StatusCode != http.StatusOK {
				respBytes, err := ioutil.ReadAll(resp.Body)
				defer resp.Body.Close()
				if err != nil {
					t.Errorf("reading from response body failed: %v ", err)
				}
				if !strings.Contains(string(respBytes), "error") {
					t.Error("expected an error key in the body json struct but found none")
				}

			}
		})
	}
}
