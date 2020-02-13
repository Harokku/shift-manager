package gsuite

import "google.golang.org/api/sheets/v4"

type License struct {
}

func (l License) Append(s *Service, r string, data [][]interface{}) (int, error) {
	var (
		values sheets.ValueRange
		err    error
	)

	values.Values = data

	res, err := s.srv.Spreadsheets.Values.Append(s.sheetId, r, &values).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return 0, err
	}

	return res.HTTPStatusCode, nil
}
