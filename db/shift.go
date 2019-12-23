package db

import (
	"database/sql"
	"errors"
	"fmt"
)

type Shift struct {
	service Service
	Id      string `json:"id"`
	Name    string `json:"name"`
	Order   int    `json:"order"`
}

func (s *Shift) New(service Service) {
	s.service = service
}

func (s *Shift) Get(name string) error {
	sqlStatement := `SELECT id,name,"order" FROM shifts WHERE name = $1`
	row := s.service.Db.QueryRow(sqlStatement, name)
	switch err := row.Scan(&s.Id, &s.Name, &s.Order); err {
	case sql.ErrNoRows:
		return errors.New("no row where retrieved")
	case nil:
		return nil
	default:
		return errors.New(fmt.Sprintf("error retrieving shift from database: %v\n", err))
	}
}

func (s *Shift) GetAll(dest *[]Shift) error {
	sqlStatement := `SELECT id,name,"order" FROM shifts`
	rows, err := s.service.Db.Query(sqlStatement)
	if err != nil {
		return errors.New(fmt.Sprintf("error retrieving shifts: %v\n", err))
	}
	defer rows.Close()

	for rows.Next() {
		var shift Shift
		err = rows.Scan(&shift.Id, &shift.Name, &shift.Order)
		if err != nil {
			return errors.New(fmt.Sprintf("error scanning row: %v\n", err))
		}
		*dest = append(*dest, shift)
	}
	err = rows.Err()
	if err != nil {
		return errors.New(fmt.Sprintf("error appending rows to result %v\n", err))
	}
	return nil
}
