package db

import "fmt"

type Coordinate struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

func (c Coordinate) String() string {
	return fmt.Sprintf("%f,%f", c.Latitude, c.Longitude)
}
