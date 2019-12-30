package api

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"net/http"
	"os"
	"shift-manager/gsuite"
)

func GetPostedShifts() echo.HandlerFunc {
	return func(context echo.Context) error {
		// Read operator name JWT
		user := context.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		n := claims["opname"].(string)
		// Sheet shift service
		s := gsuite.Service{}
		err := s.New(os.Getenv("SHEET_ID"))
		if err != nil {
			fmt.Printf("Error creating gsuite service: %v\n", err)
			return context.String(http.StatusInternalServerError, fmt.Sprintf("Error creating gsuite service: %v\n", err))
		}

		// Get all operator's posted shifts
		res, err := s.GetOperatorPostedShifts(n)
		if err != nil {
			fmt.Printf("No past shifts found: %v\n", err)
			return context.String(http.StatusNotFound, fmt.Sprintf("No past shifts found: %v", err))
		}

		// Return last 31 posted shifts
		var lastMonth []interface{}
		if len(res) > 31 {
			lastMonth = res[len(res)-31:]
		} else {
			lastMonth = res
		}
		return context.JSON(http.StatusOK, lastMonth)
	}
}
