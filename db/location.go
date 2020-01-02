package db

import (
	"database/sql"
	"errors"
	"fmt"
)

type Location struct {
	service Service
	Id      string     `json:"id"`
	Name    string     `json:"name"`
	Geo     Coordinate `json:"geo"`
	Address string     `json:"address"`
	Order   int        `json:"order"`
}

func (l *Location) New(s Service) {
	l.service = s
}

func (l *Location) Get(name string) error {
	sqlStatement := `SELECT id, name, geo[0],geo[1], address, "order" FROM locations WHERE name = $1`
	row := l.service.Db.QueryRow(sqlStatement, name)
	switch err := row.Scan(&l.Id, &l.Name, &l.Geo.Latitude, &l.Geo.Longitude, &l.Address, &l.Order); err {
	case sql.ErrNoRows:
		return errors.New("no row where retrieved")
	case nil:
		return nil
	default:
		return errors.New(fmt.Sprintf("error retrieving location from database: %v\n", err))
	}
}

func (l *Location) GetAll(dest *[]Location) error {
	sqlStatement := `SELECT id, name, geo[0],geo[1], address, "order" FROM locations`
	rows, err := l.service.Db.Query(sqlStatement)
	if err != nil {
		return errors.New(fmt.Sprintf("error retrieving locations: %v\n", err))
	}
	defer rows.Close()

	for rows.Next() {
		var location Location
		err = rows.Scan(&location.Id, &location.Name, &location.Geo.Latitude, &location.Geo.Longitude, &location.Address, &location.Order)
		if err != nil {
			return errors.New(fmt.Sprintf("error scanning row: %v\n", err))
		}
		*dest = append(*dest, location)
	}
	err = rows.Err()
	if err != nil {
		return errors.New(fmt.Sprintf("error appending rows to result %v\n", err))
	}

	return nil
}
