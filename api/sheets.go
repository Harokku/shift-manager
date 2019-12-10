package api

import (
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"shift-manager/gsuite"
	"time"
)

type shift struct {
	Timestamp         time.Time `json:"timestamp"`
	ManualCompilation bool      `json:"manual_compilation"`
	Name              string    `json:"name"`
	Date              time.Time `json:"date"`
	Shift             string    `json:"shift"`
	Vehicle           string    `json:"vehicle"`
	Role              string    `json:"role"`
	DidOverwork       bool      `json:"did_overwork"`
	OverworkEnd       time.Time `json:"overwork_end"`
	Mission           string    `json:"mission"`
	StampForgot       bool      `json:"stamp_forgot"`
	ShiftStart        time.Time `json:"shift_start"`
	ShiftEnd          time.Time `json:"shift_end"`
}

func PostShift() echo.HandlerFunc {
	return func(context echo.Context) error {
		sheetService := gsuite.Service{}
		sheetService.New()

		var s shift
		// Add post timestamp
		s.Timestamp = time.Now()

		if err := context.Bind(&s); err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error binding request body: %v", err))
		}

		var d [][]interface{}
		d = append(d, s.marshal())
		_, err := sheetService.Append("Cartellini!A4", d)
		if err != nil {
			context.String(http.StatusBadRequest, fmt.Sprintf("Error posting data to Google sheet: %v\n", err))
		}
		return context.String(http.StatusCreated, "Succesfully posted data to Google sheet")
	}
}

// preparePost prepare shift according to flags, setting null fields and default values if needed
func (s *shift) preparePost() error {
	return nil
}

// Marshal encode the struct as gsheet Value type ready to be posted
// Set null field as blank string
func (s shift) marshal() []interface{} {
	var i []interface{}

	// Append non nullable fields
	i = append(i, s.Timestamp, s.Name, s.Date, s.Shift, s.Vehicle, s.Role)

	// If DidOverwork is false, set to blank string
	if s.DidOverwork {
		i = append(i, s.DidOverwork, s.OverworkEnd, s.Mission)
	} else {
		i = append(i, s.DidOverwork, "", "")
	}

	// If StampForgot is false, set to blank string
	if s.StampForgot {
		i = append(i, s.StampForgot, s.ShiftStart, s.ShiftEnd)
	} else {
		i = append(i, s.StampForgot, "", "")
	}

	fmt.Printf("New interface: %v\n", i)

	return i
}
