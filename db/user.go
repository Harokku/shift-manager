package db

import (
	"database/sql"
	"errors"
	"fmt"
)

type User struct {
	Id       string   `json:"id"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

func (u *User) GetUser(s *Service, username string) error {
	sqlStatement := `SELECT username, password FROM users WHERE username=$1`
	row := s.Db.QueryRow(sqlStatement, username)
	switch err := row.Scan(&u.Username, &u.Password); err {
	case sql.ErrNoRows:
		return errors.New("no row where retrieved")
	case nil:
		return nil
	default:
		return errors.New(fmt.Sprintf("error retrieving user from database: %v\n", err))
	}
}
