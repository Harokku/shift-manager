package db

import (
	"database/sql"
	"errors"
	"fmt"
)

type Vehicle struct {
	service Service
	Id      string `json:"id"`
	Name    string `json:"name"`
	Order   int    `json:"order"`
}

func (v *Vehicle) New(s Service) {
	v.service = s
}

func (v *Vehicle) Get(name string) error {
	sqlStatement := `SELECT id, name, "order" FROM vehicles WHERE name = $1`
	row := v.service.Db.QueryRow(sqlStatement, name)
	switch err := row.Scan(&v.Id, &v.Name, &v.Order); err {
	case sql.ErrNoRows:
		return errors.New("no row where retrieved")
	case nil:
		return nil
	default:
		return errors.New(fmt.Sprintf("error retrieving vehicle from database: %v\n", err))
	}
}

func (v *Vehicle) GetAll(dest *[]Vehicle) error {
	sqlStatement := `SELECT id,name,"order" FROM vehicles`
	rows, err := v.service.Db.Query(sqlStatement)
	if err != nil {
		return errors.New(fmt.Sprintf("error retrieving vehicles: %v\n", err))
	}
	defer rows.Close()

	for rows.Next() {
		var vehicle Vehicle
		err = rows.Scan(&vehicle.Id, &vehicle.Name, &vehicle.Order)
		if err != nil {
			return errors.New(fmt.Sprintf("error scanning row: %v\n", err))
		}
		*dest = append(*dest, vehicle)
	}
	err = rows.Err()
	if err != nil {
		return errors.New(fmt.Sprintf("error appending rows to result %v\n", err))
	}
	return nil
}
