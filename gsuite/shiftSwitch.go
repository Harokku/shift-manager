package gsuite

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ShiftsToSwitch struct {
	service     Service         //gsheet service
	dayCoord    DayCoord        //gsheet day coordinates for lookups
	FirstName   string          `json:"first_name"` //1st operator name (as on gsheet)
	FirstDate   time.Time       `json:"first_date"` //1st operator date to change
	firstDay    [][]interface{} //Day readed from gsheet to search for coordinates
	firstCoord  string          //1st operator coordinates in gsheet !A1 format
	SecondName  string          `json:"second_name"` //2nd operator name (as on gsheet)
	SecondDate  time.Time       `json:"second_date"` //2nd operator date to change
	secondDay   [][]interface{} //Day readed from gsheet to search for coordinates
	secondCoord string          //1st operator coordinates in gsheet !A1 format
}

// New - instantiate new shifts to switch
//
// Placeholder for future initiator logic, actually only ser struct service field and initialize dayCoord
func (s *ShiftsToSwitch) New(service Service) error {
	s.service = service
	err := s.dayCoord.New()
	if err != nil {
		return errors.New(fmt.Sprintf("error retrieving day coordinates: %v\n", err))
	}
	return nil
}

// getDays populate worked day for 1st and 2nd operator based on shift fields
//
// Used to later search for coordinates to switch
func (s *ShiftsToSwitch) getDays() error {
	var err error

	// Check if all struct fields are actually populated
	if s.FirstName == "" || s.FirstDate.IsZero() || s.SecondName == "" || s.SecondDate.IsZero() {
		return errors.New(fmt.Sprintf(
			"Not all required fields supplied: 1sOperator: %v %v - 2ndOperator %v %v",
			s.FirstName,
			s.FirstDate,
			s.SecondName,
			s.SecondDate,
		))
	}

	// Get 1st operator workday from gsheet
	s.firstDay, err = s.service.ReadDay(s.dayCoord, s.FirstDate)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot retrieve 1st operator workday: %v\n", err))
	}

	// Get 2nd operator workday from gsheet
	s.secondDay, err = s.service.ReadDay(s.dayCoord, s.SecondDate)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot retrieve 2nd operator workday: %v\n", err))
	}

	return nil
}

// getCoordinates search for shift coordinate and populate 1st and 2nd operator fields, getting shifts ready to be switched.
// Method will do necessary matrix math to adapt coordinates based on DayCoord offset.
func (s *ShiftsToSwitch) getCoordinates() error {
	var err error

	// Check if all days and operators name are populated
	if s.firstDay == nil || s.FirstName == "" || s.secondDay == nil || s.SecondName == "" {
		return errors.New(fmt.Sprintf(
			"Not all required fields supplied: 1sOperator: %v %v - 2ndOperator %v %v",
			s.FirstName,
			s.firstDay,
			s.SecondName,
			s.secondDay,
		))
	}

	// Retrieve gsheet coordinates for 1st operator
	s.firstCoord, err = s.service.GetCellRange(s.firstDay, s.FirstName)
	if err != nil {
		return errors.New(fmt.Sprintf("Error retrieveing 1st operator coordinates: %v\n", err))
	}
	// Calculate real coordinate offsetting from daycoord
	s.firstCoord = offsetCoordinates(s.dayCoord, s.FirstDate, s.firstCoord)

	// Retrieve gsheet coordinates for 2nd operator
	s.secondCoord, err = s.service.GetCellRange(s.secondDay, s.SecondName)
	if err != nil {
		return errors.New(fmt.Sprintf("Error retrieveing 2nd operator coordinates: %v\n", err))
	}

	// Calculate real coordinate offsetting from daycoord
	s.secondCoord = offsetCoordinates(s.dayCoord, s.SecondDate, s.secondCoord)

	return nil
}

// SwitchShifts - actually switch shifts (will call getDays and GetCoordinates)
func (s *ShiftsToSwitch) SwitchShifts() error {
	var err error

	// Retrieve days and populate struct
	err = s.getDays()
	if err != nil {
		return errors.New(fmt.Sprintf("error getting working days: %v\n", err))
	}

	// Retrieve coordinates and populate struct
	err = s.getCoordinates()
	if err != nil {
		return errors.New(fmt.Sprintf("error getting coordinates: %v\n", err))
	}

	// -------------------
	// Actually switch selected shifts
	// -------------------

	// data represent the modified cells
	data := []CellToUpdate{
		{Range: s.firstCoord, Value: s.SecondName},
		{Range: s.secondCoord, Value: s.FirstName},
	}

	// Call method to actually update gsheet
	err = s.service.BatchUpdateCells(data)
	if err != nil {
		return errors.New(fmt.Sprintf("error swirching shifts: %v\n", err))
	}

	return nil
}

// offsetCoordinates offset dayCoord with passed coordinate
//
// c DayCoord: Day coordinates to offset
//
// d time.Time: Date to search for starting coordinate
//
// s string: Coordinates to offset
func offsetCoordinates(c DayCoord, d time.Time, s string) string {
	// Etract week from passed time (will be used to compose ghseet cell coordiantes)
	_, week := d.ISOWeek()
	// Extract day number from passed time
	day := d.Weekday()

	var startIndex string
	// Switch week day and set start index (A9:F12 -> A9)
	switch day {
	case 0:
		startIndex = strings.Split(c.sunday, ":")[0]
	case 1:
		startIndex = strings.Split(c.monday, ":")[0]
	case 2:
		startIndex = strings.Split(c.tuesday, ":")[0]
	case 3:
		startIndex = strings.Split(c.wednesday, ":")[0]
	case 4:
		startIndex = strings.Split(c.thursday, ":")[0]
	case 5:
		startIndex = strings.Split(c.friday, ":")[0]
	case 6:
		startIndex = strings.Split(c.saturday, ":")[0]
	}

	// Extract row and column index
	rowReg := regexp.MustCompile("[0-9]+")
	colReg := regexp.MustCompile("[A-Z]+")
	row, _ := strconv.Atoi(rowReg.FindString(startIndex))
	offsetRow, _ := strconv.Atoi(rowReg.FindString(s))
	col := colReg.FindString(startIndex)
	offsetCol := colReg.FindString(s)

	resRow := row + offsetRow - 1
	resCol := string([]rune(col)[0] + []rune(offsetCol)[0] - 'A')

	return fmt.Sprintf("%v!%v%v", week, resCol, resRow)
}
