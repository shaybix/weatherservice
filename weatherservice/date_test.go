package main

import (
	"testing"
)

func TestISO8601ToTime(t *testing.T) {

	_, err := ISO8601ToTime("2018-08-03T00:00:00Z")
	if err != nil {
		t.Errorf("Parsing date ISO8601 to time.Time failled with err: %v", err)
	}

}
