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

type illness struct {
	Timestamp      time.Time `json:"timestamp"`
	Name           string    `json:"name"`
	From           time.Time `json:"from"`
	To             time.Time `json:"to"`
	ProtocolNumber string    `json:"protocol_number"`
}

// PostIllness post new illness request reading from request body.
// Timestamp will be added at post.
// Name will be populated from logged in user
//
// request body:
//
// {
//		"from":	"2019-12-30T00:00:00+01:00"	// From date
//		"to":	"2019-12-30T00:00:00+01:00"	// To date
//		"protocol_number": "12345A"			// Illness certification protocol number
// }
func PostIllness() echo.HandlerFunc {
	return func(context echo.Context) error {
		var (
			sheetService gsuite.Service
			err          error
			i            illness
		)

		// Create gsheet service reading from env variable
		err = sheetService.New(os.Getenv("SHEET_ID"))
		if err != nil {
			return context.String(http.StatusInternalServerError, fmt.Sprintf("Error creating gsheet service, :%v\n", err))
		}

		// Add post timestamp
		i.Timestamp = time.Now()

		// Bind request body to license struct
		if err = context.Bind(&i); err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error binding request body: %v\n", err))
		}

		// Read operator name from auth token and set struct Name field
		user := context.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		operatorName := claims["opname"].(string)
		i.Name = operatorName

		// d is data casted and ready to be appended to google sheet
		var d [][]interface{}
		d = append(d, i.marshalGSheet())

		// Call gsheet api to append data
		_, err = sheetService.Append("Malattie!A4", d)
		if err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error posting data do Google sheet: %v\n", err))
		}

		return context.String(http.StatusCreated, "Succesfully posted permission request to Google sheets")
	}
}

// Marshal encode the struct as gsheet Value type ready to be posted
// Set null field as blank string
func (il illness) marshalGSheet() []interface{} {
	dateLayout := "02-01-2006"
	var i []interface{}

	// Append non nullable fields
	i = append(i, il.Timestamp.Format(dateLayout), il.Name, il.From.Format(dateLayout), il.To.Format(dateLayout), il.ProtocolNumber)

	return i
}
