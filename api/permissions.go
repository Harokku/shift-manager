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

type permission struct {
	Timestamp  time.Time `json:"timestamp"`
	Name       string    `json:"name"`
	Date       time.Time `json:"date"`
	From       time.Time `json:"from"`
	To         time.Time `json:"to"`
	Motivation string    `json:"motivation"`
}

// PostPermission post new permission request reading from request body.
// Timestamp will be added at post.
// Name will be populated from logged in user
//
// request body:
//
// {
//		"date":	"2019-12-30T00:00:00+01:00"	// Shift to request permission from
//		"from":	"2019-12-30T00:00:00+01:00"	// From time
//		"to":	"2019-12-30T00:00:00+01:00"	// To Time
//		"motivation": "I have to"			// Motivation to ask for a permission
// }
func PostPermission() echo.HandlerFunc {
	return func(context echo.Context) error {
		var (
			sheetService gsuite.Service
			err          error
			p            permission
		)

		// Create gsheet service reading from env variable
		err = sheetService.New(os.Getenv("SHEET_ID"))
		if err != nil {
			return context.String(http.StatusInternalServerError, fmt.Sprintf("Error creating gsheet service, :%v\n", err))
		}

		// Add post timestamp
		p.Timestamp = time.Now()

		// Bind request body to license struct
		if err = context.Bind(&p); err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error binding request body: %v\n", err))
		}

		// Read operator name from auth token and set struct Name field
		user := context.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		operatorName := claims["opname"].(string)
		p.Name = operatorName

		// d is data casted and ready to be appended to google sheet
		var d [][]interface{}
		d = append(d, p.marshalGSheet())

		// Call gsheet api to append data
		_, err = sheetService.Append("PermessiOrari!A4", d)
		if err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error posting data do Google sheet: %v\n", err))
		}

		return context.String(http.StatusCreated, "Succesfully posted permission request to Google sheets")
	}
}

// Marshal encode the struct as gsheet Value type ready to be posted
// Set null field as blank string
func (p permission) marshalGSheet() []interface{} {
	dateLayout := "02-01-2006"
	timeLayout := "15:04"
	var i []interface{}

	// Append non nullable fields
	i = append(i, p.Timestamp.Format(dateLayout), p.Name, p.Date.Format(dateLayout), p.From.Format(timeLayout), p.To.Format(timeLayout), p.Motivation)

	return i
}
