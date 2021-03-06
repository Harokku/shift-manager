package api

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"net/http"
	"os"
	"shift-manager/gsuite"
	"strings"
	"time"
)

type shift struct {
	Timestamp         time.Time `json:"timestamp"`
	ManualCompilation bool      `json:"manual_compilation"`
	Motivation        string    `json:"motivation"`
	Name              string    `json:"name"`
	Date              time.Time `json:"date"`
	Location          string    `json:"location"`
	Shift             string    `json:"shift"`
	Vehicle           string    `json:"vehicle"`
	Role              string    `json:"role"`
	Note              string    `json:"note"`
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
		err := sheetService.New(os.Getenv("SHEET_ID"))
		if err != nil {
			return context.String(http.StatusInternalServerError, fmt.Sprintf("Error creating gSheet service: %v\n", err))
		}

		var s shift
		// Add post timestamp
		s.Timestamp = time.Now()

		// Bind request body to shift struct
		if err := context.Bind(&s); err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error binding request body: %v\n", err))
		}

		// Read operator name from auth token and set struct Name field
		user := context.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		operatorName := claims["opname"].(string)
		s.Name = operatorName

		err = s.setDefaults(strings.Split(operatorName, " ")[0])
		if err != nil {
			fmt.Printf("Cannot retrieve assigned shift data, falling back to declared: %v\n", err)
		}

		// d is data casted and ready to be appended to google sheet
		var d [][]interface{}
		d = append(d, s.marshalGSheet())
		_, err = sheetService.Append("Cartellini!A4", d)
		if err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error posting data to Google sheet: %v\n", err))
		}
		return context.String(http.StatusCreated, "Succesfully posted data to Google sheet")
	}
}

// setDefaults retrieve default shift from gsheet and set default value
func (s *shift) setDefaults(name string) error {
	if s.ManualCompilation {
		return nil
	}

	// Retrieve day coordinates
	dayCoord := gsuite.DayCoord{}
	dayCoord.New()

	// Gsheet service
	srv := gsuite.Service{}
	srv.New(os.Getenv("SHIFT_ID"))

	// Retrive today shift
	todayShift, err := srv.ReadDay(dayCoord, s.Date)
	if err != nil {
		fmt.Printf("Cannot retrieve requested shift %v\n", err)
		return errors.New("cannot retrieve requested shift")
	}

	// Retrieve today role
	todayRoles, err := srv.GetOperatorRoles(todayShift, name)
	if err != nil {
		fmt.Printf("Cannot retrieve requested shoft roles, operator not found: %v\n", err)
		return errors.New("cannot retrieve requested shoft roles, operator not found")
	}

	// Populate roles
	splitRoles := strings.Split(todayRoles, "|")
	s.Location = splitRoles[0]
	s.Shift = splitRoles[1]
	s.Vehicle = splitRoles[2]
	s.Role = splitRoles[3]

	return nil
}

// Marshal encode the struct as gsheet Value type ready to be posted
// Set null field as blank string
func (s shift) marshalGSheet() []interface{} {
	dateLayout := "02-01-2006"
	timeLayout := "15:04"
	var i []interface{}

	// Append non nullable fields
	i = append(i, s.Timestamp.Format(dateLayout), s.Name, s.Date.Format(dateLayout), s.ManualCompilation, s.Motivation, s.Location, s.Shift, s.Vehicle, s.Role, s.Note)

	// If DidOverwork is false, set to blank string
	if s.DidOverwork {
		i = append(i, s.DidOverwork, s.OverworkEnd.Format(timeLayout), s.Mission)
	} else {
		i = append(i, s.DidOverwork, "", "")
	}

	// If StampForgot is false, set to blank string
	if s.StampForgot {
		i = append(i, s.StampForgot, s.ShiftStart.Format(timeLayout), s.ShiftEnd.Format(timeLayout))
	} else {
		i = append(i, s.StampForgot, "", "")
	}

	return i
}
