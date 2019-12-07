package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
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

func TestUser_CreateUser(t *testing.T) {
	dbConn, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	checkErrorAndPanic(err)

	defer dbConn.Close()

	// Test credential
	username := "testCreate"
	password := "test"

	// User DB service creation
	user := User{}
	db := Service{Db: dbConn}
	user.New(db)
	err = user.CreateUser(username, password)
	if err != nil {
		t.Errorf("Error occurred while creating new user: %v\n", err)
	}
	fmt.Printf("Created new user: %v\n", user)
	err = user.DeleteUser(username)
	if err != nil {
		t.Errorf("Error occurred whiel deleting test user: %v - Manual intervention required to clean user: testCreate \n", err)
	}
}
