package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"os"
	"reflect"
	"testing"
)

func TestUser_GetUser(t *testing.T) {
	dbConn, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	checkErrorAndPanic(err)

	defer dbConn.Close()

	user := User{}
	expectedUser := User{
		Username: "test",
		Password: "$2a$04$vRlRz0WiJVvP13k4boYY3eA2Ye8OOTyixWFFYiL.eACzvX2Z5JEBm",
	}
	db := Service{Db: dbConn}
	user.New(db)
	err = user.GetUser("test")
	checkErrorAndPanic(err)

	fieldsToCheck := User{
		Username: user.Username,
		Password: user.Password,
	}
	if reflect.DeepEqual(fieldsToCheck, expectedUser) == false {
		t.Errorf("retrieved user mismatch got: %v - expected %v\n", user, expectedUser)
	}
}

// Default error check with fatal if err != nil
func checkErrorAndPanic(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
