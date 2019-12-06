package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
)

type User struct {
	service  Service
	Id       string   `json:"id"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

func (u *User) New(s Service) {
	u.service = s
}

func (u *User) GetUser(username string) error {
	sqlStatement := `SELECT username, password,roles FROM users WHERE username=$1`
	row := u.service.Db.QueryRow(sqlStatement, username)
	switch err := row.Scan(&u.Username, &u.Password, pq.Array(&u.Roles)); err {
	case sql.ErrNoRows:
		return errors.New("no row where retrieved")
	case nil:
		return nil
	default:
		return errors.New(fmt.Sprintf("error retrieving user from database: %v\n", err))
	}
}
