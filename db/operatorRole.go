package db

import (
	"database/sql"
	"errors"
	"fmt"
)

type OperatorRole struct {
	service Service
	Id      string `json:"id"`
	Name    string `json:"name"`
	Order   int    `json:"order"`
}

func (r *OperatorRole) New(service Service) {
	r.service = service
}

func (r *OperatorRole) Get(name string) error {
	sqlStatement := `SELECT id, name, "order" FROM operator_roles WHERE name = $1`
	row := r.service.Db.QueryRow(sqlStatement, name)
	switch err := row.Scan(&r.Id, &r.Name, &r.Order); err {
	case sql.ErrNoRows:
		return errors.New("no row where retrieved")
	case nil:
		return nil
	default:
		return errors.New(fmt.Sprintf("error retrieving role from database: %v\n", err))
	}
}

func (r *OperatorRole) GetAll(dest *[]OperatorRole) error {
	sqlStatement := `SELECT id,name,"order" FROM operator_roles`
	rows, err := r.service.Db.Query(sqlStatement)
	if err != nil {
		return errors.New(fmt.Sprintf("error retrieving roles: %v\n", err))
	}
	defer rows.Close()

	for rows.Next() {
		var roles OperatorRole
		err = rows.Scan(&roles.Id, &roles.Name, &roles.Order)
		if err != nil {
			return errors.New(fmt.Sprintf("error scanning row: %v\n", err))
		}
		*dest = append(*dest, roles)
	}
	err = rows.Err()
	if err != nil {
		return errors.New(fmt.Sprintf("error appending rows to result %v\n", err))
	}
	return nil
}
