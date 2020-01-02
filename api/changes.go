package api

import (
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"os"
	"shift-manager/gsuite"
	"time"
)

type change struct {
	FirstDate  time.Time `json:"first_date"`
	FirstName  string    `json:"first_name"`
	SecondDate time.Time `json:"second_date"`
	SecondName string    `json:"second_name"`
}

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
		return context.String(http.StatusNoContent, "Shift correctly modified")
	}
}
