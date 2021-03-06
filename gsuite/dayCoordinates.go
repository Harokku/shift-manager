package gsuite

import (
	"errors"
	"fmt"
	"log"
	"os"
)

type DayCoord struct {
	sheetId   string
	monday    string
	tuesday   string
	wednesday string
	thursday  string
	friday    string
	saturday  string
	sunday    string
}

// Initialize a new gsheet day coordinates struct reading from gsheet range
// r string: the range where read data from in the !A1 format (Sheet!A1:B2)
func (c *DayCoord) New() error {
	// Create new gsheet service, passing config sheetId id
	service := Service{}
	c.sheetId = os.Getenv("SHIFT_ID")
	err := service.New(c.sheetId)
	if err != nil {
		return errors.New(fmt.Sprintf("error creating gsheet service: %v\n", err))
	}

	// Call read method to actually retrieve data
	response, err := service.ReadRange(os.Getenv("WEEKDAY_RANGE"))
	if err != nil {
		return errors.New(fmt.Sprintf("error retrieving data from gsheet: %v\n", err))
	}

	// Cycle through response populating struct
	c.monday = fmt.Sprintf("%s:%s", response[0][0], response[0][1])
	c.tuesday = fmt.Sprintf("%s:%s", response[1][0], response[1][1])
	c.wednesday = fmt.Sprintf("%s:%s", response[2][0], response[2][1])
	c.thursday = fmt.Sprintf("%s:%s", response[3][0], response[3][1])
	c.friday = fmt.Sprintf("%s:%s", response[4][0], response[4][1])
	c.saturday = fmt.Sprintf("%s:%s", response[5][0], response[5][1])
	c.sunday = fmt.Sprintf("%s:%s", response[6][0], response[6][1])
	return nil
}

// Update day coordinates from passed 2d array of string
func (c *DayCoord) Update(coord [][]string) error {
	// Cycle through coord param populating struct
	c.monday = fmt.Sprintf("%s:%s", coord[0][0], coord[0][1])
	c.tuesday = fmt.Sprintf("%s:%s", coord[1][0], coord[1][1])
	c.wednesday = fmt.Sprintf("%s:%s", coord[2][0], coord[2][1])
	c.thursday = fmt.Sprintf("%s:%s", coord[3][0], coord[3][1])
	c.friday = fmt.Sprintf("%s:%s", coord[4][0], coord[4][1])
	c.saturday = fmt.Sprintf("%s:%s", coord[5][0], coord[5][1])
	c.sunday = fmt.Sprintf("%s:%s", coord[6][0], coord[6][1])

	return nil
}

// Default error check with fatal if err != nil
func CheckErrorAndPanic(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
