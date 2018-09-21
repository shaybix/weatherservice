package main

import (
	"fmt"
	"time"
)

// ISO8601ToTime converts a ISO8601 string to type time.Time
func ISO8601ToTime(date string) (time.Time, error) {
	var t time.Time
	if date == "" {
		return t, fmt.Errorf("query param is empty")
	}

	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return t, fmt.Errorf("query param contains invalid ISO8601 date")
	}
	if !time.Now().After(t) {
		return t, fmt.Errorf("query params contains date in the future")
	}
	return t, nil
}
