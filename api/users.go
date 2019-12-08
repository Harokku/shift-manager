package api

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"net/http"
	"shift-manager/db"
)

func Login(s *db.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		user := struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{}

		if err := context.Bind(&user); err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error binding request body: %v", err))
		}

		t, err := db.CreateToken(user.Username, user.Password, *s)
		if err != nil {
			return context.String(http.StatusUnauthorized, "Error authenticating user")
		}
		return context.JSON(http.StatusOK, map[string]string{
			"token": t,
		})
	}
}

func ResetPwd(s *db.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		// Read user from JWT and extract claims
		user := context.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		roles := claims["role"].([]interface{})

		// Username and password passed in as JSON POST body
		r := struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{}

		if err := context.Bind(&r); err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error binding request body: %v\n", err))
		}

		// Call db API to reset password
		u := db.User{}
		u.New(*s)

		if checkIfAdmin(roles) {
			err := u.ResetPassword(r.Username, r.Password)
			if err != nil {
				return context.String(http.StatusBadRequest, fmt.Sprintf("Error resetting password for user %v: %v", r.Username, err))
			}
			return context.String(http.StatusOK, fmt.Sprintf("Password reset for user %v", u.Username))
		}
		return context.String(http.StatusUnauthorized, "User is not admin, can't reset password")
	}
}

// Check if the passed claim contain "admin" role
// i: jwt role claim
func checkIfAdmin(i []interface{}) bool {
	for _, role := range i {
		if role == "admin" {
			return true
		}
	}
	return false
}
