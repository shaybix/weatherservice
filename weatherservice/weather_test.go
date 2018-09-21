package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWeatherHandler(t *testing.T) {
	tt := []struct {
		name      string
		startDate string
		endDate   string
		status    int
	}{
		{"invalid start and end date", "Monday 16 September 2017", "Wednesday 18 September 2017", http.StatusBadRequest},
		{"invalid end date", "2018-08-03T00:00:00Z", "Wednesday 18 September 2017", http.StatusBadRequest},
		{"empty end date", "2018-08-03T00:00:00Z", "", http.StatusBadRequest},
		{"empty start date", "", "2018-08-03T00:00:00Z", http.StatusBadRequest},
		{"valid dates with status 200", "2018-05-10T00:00:00Z", "2018-05-15T00:00:00Z", http.StatusOK},
		{"dates in the future", "2018-12-10T00:00:00Z", "2018-12-18T00:00:00Z", http.StatusBadRequest},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, tc := range tt {
		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/?start=%s&end=%s", tc.startDate, tc.endDate), nil)
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

	}
}
