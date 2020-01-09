package api

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"net/http"
	"os"
	"shift-manager/db"
	"shift-manager/gsuite"
	"time"
)

type change struct {
	FirstDate  time.Time `json:"first_date"`
	FirstName  string    `json:"first_name"`
	SecondDate time.Time `json:"second_date"`
	SecondName string    `json:"second_name"`
}

// PutChange actually modify gsheet shift table switching passed operators
//
// Request body:
// {
//		first_date: Requester date
//		first_name: Requester operator name
//		second_date: Requested date
//		second_name: Requested operator name
// }
func PutChange() echo.HandlerFunc {
	return func(context echo.Context) error {
		var err error
		sheetService := gsuite.Service{}
		err = sheetService.New(os.Getenv("SHIFT_ID"))
		if err != nil {
			fmt.Printf("Error creating gSheet service: %v\n", err)
			return context.String(http.StatusInternalServerError, fmt.Sprintf("Error creating gSheet service: %v\n", err))
		}

		var c change

		// Bind request body to change struct
		if err := context.Bind(&c); err != nil {
			fmt.Printf("Error binding request body: %v\n", err)
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error binding request body: %v\n", err))
		}

		// Create an instance of gsuite.ShiftToSwitch and populate fields from request body
		sc := gsuite.ShiftsToSwitch{}
		err = sc.New(sheetService)
		if err != nil {
			fmt.Printf("Error creating shift change service: %v\n", err)
			return context.String(http.StatusInternalServerError, fmt.Sprintf("Error creating shift change service: %v\n", err))
		}

		sc.FirstDate = c.FirstDate
		sc.FirstName = c.FirstName
		sc.SecondDate = c.SecondDate
		sc.SecondName = c.SecondName

		// Call service to actually modify gsheet
		err = sc.SwitchShifts()
		if err != nil {
			fmt.Printf("Error switching shifts: %v,\n", err)
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error switching shifts: %v,\n", err))
		}
		return context.String(http.StatusOK, "Shift correctly modified")
	}
}

// RequestChange create a new shift change request to DB. Will be posted to gsheet after is been managed
//
// Request body:
// {
//		applicant_date: Requester date
//		with_date: Requested date
//		with_name: Requested operator name
// }
func RequestChange(s *db.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		var (
			err         error
			requester   db.User        // Requester user data (username and ID)
			shiftChange db.ShiftChange // Shift change service
		)

		// Read user from JWT and extract claims
		user := context.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		username := claims["username"].(string)

		// Create service and get logged in user's DB ID
		requester.New(*s)
		err = requester.GetUser(username)

		// Populate shiftChange with request body data
		shiftChange.New(*s)
		if err = context.Bind(&shiftChange); err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("error binding request body: %v\n", err))
		}
		shiftChange.ApplicantName = requester.Id
		err = shiftChange.NewRequest()
		if err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("error creating shift request: %v\n", err))
		}

		return context.String(http.StatusOK, "Shift change request correctly submitted")
	}
}
