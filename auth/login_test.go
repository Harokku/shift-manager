package auth

import (
	"database/sql"
	"log"
	"os"
	"shift-manager/db"
	"testing"
)

func TestLogin(t *testing.T) {
	dbConn, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	checkErrorAndPanic(err)

	defer dbConn.Close()
	database := db.Service{Db: dbConn}

	_, err = Login("test", "plinioilbasso", database)
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
