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

func GetAllUserNames(s *db.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		var (
			u     db.User
			users []db.User
			err   error
		)

		u.New(*s)

		err = u.GetAllUser(&users)
		if err != nil {
			return context.String(http.StatusNotFound, fmt.Sprintf("Error retrieving users: %v\n", err))
		}
		return context.JSON(http.StatusOK, users)
	}
}

func GetUserDetailsFromClaims(s *db.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		// Read user from JWT and extract username claim
		user := context.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		username := claims["username"].(string)

		// Call db API to get user detail
		u := db.User{}
		u.New(*s)

		err := u.GetUserDetail(username)
		if err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error retrieving user's detail: %v\n", err))
		}
		return context.JSON(http.StatusOK, u)
	}
}

// TODO: Refactor admin check using middleware (checkIfRole)
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
			return context.String(http.StatusOK, fmt.Sprintf("Password reset for user %v", r.Username))
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
