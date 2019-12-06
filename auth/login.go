package auth

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"os"
	"shift-manager/db"
	"time"
)

func Login(username, password string, s db.Service) (string, error) {
	user := db.User{}
	user.New(s)

	err := user.GetUser(username)
	if err != nil {
		return "", err
	}

	if !ComparePassword(user.Password, password) {
		return "", echo.ErrUnauthorized
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Username
	claims["role"] = user.Roles
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error signing token: %v\n", err))
	}
	return t, nil
}
