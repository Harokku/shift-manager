package api

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"net/http"
	"os"
	"shift-manager/gsuite"
	"time"
)

type license struct {
	Timestamp       time.Time `json:"timestamp"`
	Name            string    `json:"name"`
	From            time.Time `json:"from"`
	To              time.Time `json:"to"`
	With            string    `json:"with"`
	Motivation      string    `json:"motivation"`
	FromCoordinator bool      `json:"from_coordinator"`
}

// PostLicense post new license request reading from request body.
// Timestamp will be added at post.
// Name will be populated from logged in user
//
// request body:
//
// {
//		"from":	"2019-12-30T00:00:00+01:00"	// Shift to change from
//		"to":	"2019-12-30T00:00:00+01:00"	// Shift to change to
//		"motivation": "I have to"			// Motivation to ask for a change
//		"from_coordinator": true			// If change is asked from coordinator
// }
func PostLicense() echo.HandlerFunc {
	return func(context echo.Context) error {
		var (
			sheetService gsuite.Service
			err          error
			l            license
		)

		// Create gsheet service reading from env variable
		err = sheetService.New(os.Getenv("SHEET_ID"))
		if err != nil {
			return context.String(http.StatusInternalServerError, fmt.Sprintf("Error creating gsheet service, :%v\n", err))
		}

		// Add post timestamp
		l.Timestamp = time.Now()

		// Bind request body to license struct
		if err = context.Bind(&l); err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error binding request body: %v\n", err))
		}

		// Read operator name from auth token and set struct Name field
		user := context.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		operatorName := claims["opname"].(string)
		l.Name = operatorName

		// d is data casted and ready to be appended to google sheet
		var d [][]interface{}
		d = append(d, l.marshalGSheet())

		// Call gsheet api to append data
		_, err = sheetService.Append("Ferie!A4", d)
		if err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error posting data do Google sheet: %v\n", err))
		}

		return context.String(http.StatusCreated, "Succesfully posted license request to Google sheets")
	}
}

// Marshal encode the struct as gsheet Value type ready to be posted
// Set null field as blank string
func (l license) marshalGSheet() []interface{} {
	dateLayout := "02-01-2006"
	var i []interface{}

	// Append non nullable fields
	i = append(i, l.Timestamp.Format(dateLayout), l.Name, l.From.Format(dateLayout), l.To.Format(dateLayout), l.Motivation, l.FromCoordinator)

	return i
}
