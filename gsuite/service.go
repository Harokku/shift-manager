package gsuite

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"os"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	srv     *sheets.Service
	sheetId string
}

// Represent Google API service with auth and sheetId ID read from env
// GOOGLE_API is the auth secret
// SHEETS_ID is the sheetId to read from
func (s *Service) New(sheetId string) error {
	secret := os.Getenv("GOOGLE_API")
	if secret == "" {
		panic("Can't read secret from env.")
	}
	conf, err := google.JWTConfigFromJSON([]byte(secret), sheets.SpreadsheetsScope)
	CheckErrorAndPanic(err)

	srv, err := sheets.NewService(context.TODO(), option.WithHTTPClient(conf.Client(context.TODO())))
	CheckErrorAndPanic(err)

	s.srv = srv
	s.sheetId = sheetId
	return nil
}

// Append data after selected range and return the result
// r string: the range after witch append data in the !A1 format (Sheet!A1:B2)
// data [][]interface{}: 2D array with data to append
// Return int with return code (HTTPStatusCode)
func (s Service) Append(r string, data [][]interface{}) (int, error) {
	var values = sheets.ValueRange{
		Values: data,
	}

	res, err := s.srv.Spreadsheets.Values.Append(s.sheetId, r, &values).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return 0, err
	}
	return res.HTTPStatusCode, nil
}

// ReadRange read data from selected range and return it
// r string: Range to search in !A1 format
// Return [][]interface{}: retrieved data
func (s Service) ReadRange(r string) ([][]interface{}, error) {
	res, err := s.srv.Spreadsheets.Values.Get(s.sheetId, r).Do()
	if err != nil {
		return nil, err
	}
	return res.Values, nil
}

// Read a single cell, if passed a bigger range discard all but single cell and return it
func (s Service) ReadCell(r string) (string, error) {
	res, err := s.srv.Spreadsheets.Values.Get(s.sheetId, r).Do()
	if err != nil {
		return "", err
	}

	cell := res.Values[0][0]
	if cell == "" {
		return "", errors.New("no cell found")
	}
	return cell.(string), nil
}

// Read day data from GSheet based on parameters
// c DayCoord: Gsheet search ranges
// t Time: Day to retrieve
// Return [][]interface{}: requested day data
func (s Service) ReadDay(c DayCoord, t time.Time) ([][]interface{}, error) {
	_, week := t.ISOWeek() // Read weekday from passed time (will be sgheet tab reference)
	day := t.Weekday()     // Read week day from passed time (will be matched vs DayCoord for interval)
	var searchRange string

	// Switch week day and set search range
	switch day {
	case 0:
		searchRange = c.sunday
	case 1:
		searchRange = c.monday
	case 2:
		searchRange = c.tuesday
	case 3:
		searchRange = c.wednesday
	case 4:
		searchRange = c.thursday
	case 5:
		searchRange = c.friday
	case 6:
		searchRange = c.saturday
	}

	// Actually retrieve data from gsheet and return
	query := fmt.Sprintf("%s!%s", strconv.Itoa(week), searchRange)
	res, err := s.srv.Spreadsheets.Values.Get(s.sheetId, query).Do()
	if err != nil {
		return nil, err
	}

	return res.Values, nil
}

// GetOperatorRoles search for name (n) in 2D array (d) and return assigned roles for that day
// d [][]interface{}: Shift day matrix
// n string: Operator name to search for
// Return string: Pipe separated values representing operator's assigned roles
func (s Service) GetOperatorRoles(d [][]interface{}, n string) (string, error) {
	cellRange, err := s.GetCellRange(d, n)
	if err != nil {
		return "", err
	}

	// Fetch roles string from gsheet and return if found
	query := fmt.Sprintf("%s%s", strings.SplitAfter(os.Getenv("ROLES_RANGE"), "!")[0], cellRange)
	res, err := s.ReadCell(query)
	if err != nil {
		return "", err
	}

	return res, nil
}

// GetCellRange search for name (n) in 2D array (d) and return a string representing gsheets cell coordinate
// Cell coordinate are supposed starting from A1 cell, so do the necessary math if offset
func (s Service) GetCellRange(d [][]interface{}, n string) (string, error) {
	// Convert operator name to lowercase for comparison
	nLowcase := strings.ToLower(n)
	// GSheet coordinate of assigned day roles
	var rolesCell string

	// Cycle through day matrix and search for operator name, if found return gsheet coordinate
	for rowIndex, row := range d {
		//fmt.Println(rowIndex)
		//fmt.Println(row)
		for colIndex, cell := range row {
			//fmt.Println(rowIndex, colIndex)
			//fmt.Println(cell)
			if strings.ToLower(cell.(string)) == nLowcase {
				//fmt.Printf("----Found match with %s, index: %d:%d----\n", cell, rowIndex, colIndex)
				sheetRow := strconv.Itoa(rowIndex + 1)
				sheetCol := string('A' + colIndex)
				//fmt.Printf("---Sheet range %s:%s---\n", sheetCol, sheetRow)
				rolesCell = fmt.Sprintf("%s%s", sheetCol, sheetRow)
			}
		}
	}

	// If no operator found return the error
	if rolesCell == "" {
		return "", errors.New("no roles found for passed operator")
	}

	return rolesCell, nil
}
