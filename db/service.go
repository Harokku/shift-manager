package db

import (
	"database/sql"
)

type Service struct {
	Db *sql.DB
}
