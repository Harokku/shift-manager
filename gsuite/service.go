package gsuite

import (
	"context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"os"
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

// Read day data from GSheet based on parameters
// c DayCoord: Coordinates of weeks days
// w string: Week number
// d string: Day name
// Return 2D Array of string with requested day data
// TODO: implement methods
