package api

import (
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"shift-manager/db"
)

func GetLocation(s *db.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		l := db.Location{}
		l.New(*s)
		name := context.Param("name")

		err := l.Get(name)
		if err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error retireving location: %v\n", err))
		}
		return context.JSON(http.StatusOK, l)
	}
}

func GetAllLocations(s *db.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		l := db.Location{}
		l.New(*s)

		var res []db.Location
		err := l.GetAll(&res)
		if err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error retrieving locations: %v\n", err))
		}
		return context.JSON(http.StatusOK, res)
	}
}
