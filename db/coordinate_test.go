package db

import "testing"

var testCoord = Coordinate{
	Latitude:  45.796827,
	Longitude: 8.846576,
}

func TestCoordinates_String(t *testing.T) {
	got := testCoord.String()
	expected := "45.796827,8.846576"

	if got != expected {
		t.Errorf("Returned string mismatch, got:  %s  -  expected:  %s", got, expected)
	}
}
