package api

import (
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"shift-manager/db"
)

func GetAllFormData(s *db.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		var response = struct {
			Locations []db.Location     `json:"locations,omitempty"`
			Shifts    []db.Shift        `json:"shifts,omitempty"`
			Vehicle   []db.Vehicle      `json:"vehicle,omitempty"`
			Roles     []db.OperatorRole `json:"roles,omitempty"`
		}{}

		// Location service
		l := db.Location{}
		l.New(*s)
		// Shift service
		shift := db.Shift{}
		shift.New(*s)
		// Vehicle service
		v := db.Vehicle{}
		v.New(*s)
		// Operator roles service
		r := db.OperatorRole{}
		r.New(*s)

		// Locations retrieval
		var locations []db.Location
		err := l.GetAll(&locations)
		if err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error retrieving locations: %v\n", err))
		}
		response.Locations = locations

		// Shifts retrieval
		var shifts []db.Shift
		err = shift.GetAll(&shifts)
		if err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error retrieving shifts: %v\n", err))
		}
		response.Shifts = shifts

		// Vehicles retrieval
		var vehicles []db.Vehicle
		err = v.GetAll(&vehicles)
		if err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error retrieving vehicles: %v\n", err))
		}

		response.Vehicle = vehicles

		// Roles retrieval
		var roles []db.OperatorRole
		err = r.GetAll(&roles)
		if err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error retrieveing operator roles: %v\n", err))
		}

		response.Roles = roles

		return context.JSON(http.StatusOK, response)
	}
}
