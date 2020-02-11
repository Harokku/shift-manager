package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"shift-manager/auth"
)

type User struct {
	service  Service
	Id       string   `json:"id"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
	Surname  string   `json:"surname"`
	Name     string   `json:"name"`
	Mail     string   `json:"mail"`
}

func (u *User) New(s Service) {
	u.service = s
}

func (u User) GetAllUser(dest *[]User) error {
	var (
		rows *sql.Rows
		err  error
	)
	sqlStatement := `select "user", surname, name, mail
					from operators
					order by surname asc
					`
	rows, err = u.service.Db.Query(sqlStatement)
	if err != nil {
		return errors.New(fmt.Sprintf("error retrieving users: %v\n", err))
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err = rows.Scan(&user.Id, &user.Surname, &user.Name, &user.Mail)
		if err != nil {
			return errors.New(fmt.Sprintf("error scanning row: %v/n", err))
		}
		*dest = append(*dest, user)
	}
	err = rows.Err()
	if err != nil {
		fmt.Sprintf("error appending row to result: %v\n", err)
	}

	return nil
}

func (u *User) GetUser(username string) error {
	sqlStatement := `SELECT id, username, password,roles FROM users WHERE username=$1`
	row := u.service.Db.QueryRow(sqlStatement, username)
	switch err := row.Scan(&u.Id, &u.Username, &u.Password, pq.Array(&u.Roles)); err {
	case sql.ErrNoRows:
		return errors.New("no row where retrieved")
	case nil:
		return nil
	default:
		return errors.New(fmt.Sprintf("error retrieving user from database: %v\n", err))
	}
}

func (u *User) GetUserDetail(username string) error {
	sqlStatement := `SELECT
						   u.username,
						   o.surname,
						   o.name,
						   o.mail
					FROM users u
					INNER JOIN operators o on u.id = o."user"
					WHERE u.username = $1`

	row := u.service.Db.QueryRow(sqlStatement, username)
	switch err := row.Scan(&u.Username, &u.Surname, &u.Name, &u.Mail); err {
	case sql.ErrNoRows:
		return errors.New("no row where retrieved")
	case nil:
		// Doubleckeck that no password or ID field is returned explicitly setting it to null
		u.Password = ""
		u.Id = ""
		return nil
	default:
		return errors.New(fmt.Sprintf("error retrieving user from database: %v\n", err))
	}
}

func (u *User) CreateUser(username, password string) error {
	sqlStatement := `
		INSERT INTO users (username, password)
		VALUES ($1,$2)
		RETURNING id
`
	hashedPwd, err := auth.HashAndSalt(password)
	if err != nil {
		return errors.New(fmt.Sprintf("Error hashing password: %v\n", err))
	}

	err = u.service.Db.QueryRow(sqlStatement, username, hashedPwd).Scan(&u.Id)
	if err != nil {
		return errors.New(fmt.Sprintf("Error creating new user: %v\n", err))
	}
	return nil
}

func (u *User) ResetPassword(username, password string) error {
	sqlStatement := `
		UPDATE users
		SET password = $2
		WHERE username=$1
`
	hashedPwd, err := auth.HashAndSalt(password)
	if err != nil {
		return errors.New(fmt.Sprintf("Error hashing password: %v\n", err))
	}

	_, err = u.service.Db.Exec(sqlStatement, username, hashedPwd)
	if err != nil {
		return errors.New(fmt.Sprintf("Error occurred while resetting password: %v\n", err))
	}
	return nil
}

func (u *User) DeleteUser(username string) error {
	sqlStatement := `
		DELETE FROM users
		WHERE username = $1
`
	_, err := u.service.Db.Exec(sqlStatement, username)
	if err != nil {
		return errors.New(fmt.Sprintf("Error deleting user: %v\n", err))
	}
	return nil
}
