package db

import (
	"database/sql"
	"log"
	"os"
	"testing"
)

func TestLogin(t *testing.T) {
	dbConn, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	checkErrorAndPanic(err)

	defer dbConn.Close()
	database := Service{Db: dbConn}

	_, err = CreateToken("test", "plinioilbasso", database)
	if err != nil {
		t.Errorf("Token creation failed: %v\n", err)
	}
}

// Default error check with fatal if err != nil
func checkErrorAndPanic(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
