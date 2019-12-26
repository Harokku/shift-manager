package gsuite

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type DayCoord struct {
	sheetId   string
	sheetName string
	monday    string
	tuesday   string
	wednesday string
	thursday  string
	friday    string
	saturday  string
}

// Initialize a new gsheet day coordinates struct reading from gsheet range
// r string: the range where read data from in the !A1 format (Sheet!A1:B2)
func (c *DayCoord) New() error {
	// Create new gsheet service, passing config sheetId id
	service := Service{}
	c.sheetId = os.Getenv("SHIFT_ID")
	c.sheetName = strings.Split(os.Getenv("WEEKDAY_RANGE"), "!")[0]
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
	c.monday = fmt.Sprintf("%s!%s:%s", c.sheetName, response[0][0], response[0][1])
	c.tuesday = fmt.Sprintf("%s!%s:%s", c.sheetName, response[1][0], response[1][1])
	c.wednesday = fmt.Sprintf("%s!%s:%s", c.sheetName, response[2][0], response[2][1])
	c.thursday = fmt.Sprintf("%s!%s:%s", c.sheetName, response[3][0], response[3][1])
	c.friday = fmt.Sprintf("%s!%s:%s", c.sheetName, response[4][0], response[4][1])
	c.saturday = fmt.Sprintf("%s!%s:%s", c.sheetName, response[5][0], response[5][1])
	return nil
}

// Update day coordinates from passed 2d array of string
func (c *DayCoord) Update(coord [][]string) error {
	// Cycle through coord param populating struct
	c.monday = fmt.Sprintf("%s!%s:%s", c.sheetName, coord[0][0], coord[0][1])
	c.tuesday = fmt.Sprintf("%s!%s:%s", c.sheetName, coord[1][0], coord[1][1])
	c.wednesday = fmt.Sprintf("%s!%s:%s", c.sheetName, coord[2][0], coord[2][1])
	c.thursday = fmt.Sprintf("%s!%s:%s", c.sheetName, coord[3][0], coord[3][1])
	c.friday = fmt.Sprintf("%s!%s:%s", c.sheetName, coord[4][0], coord[4][1])
	c.saturday = fmt.Sprintf("%s!%s:%s", c.sheetName, coord[5][0], coord[5][1])

	return nil
}

// Default error check with fatal if err != nil
func CheckErrorAndPanic(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
