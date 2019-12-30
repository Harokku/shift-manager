package api

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"net/http"
	"os"
	"shift-manager/db"
	"shift-manager/gsuite"
	"strings"
	"time"
)

func GetAllFormData(s *db.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		var response = struct {
			Locations []db.Location     `json:"locations,omitempty"`
			Shifts    []db.Shift        `json:"shifts,omitempty"`
			Vehicles  []db.Vehicle      `json:"vehicles,omitempty"`
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

		response.Vehicles = vehicles

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

func GetLoggedInOperatorShift() echo.HandlerFunc {
	return func(context echo.Context) error {
		// Read operator name fom JWT
		user := context.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		opFullName := claims["opname"]

		// Extract operator surname from claim
		name := strings.Split(opFullName.(string), " ")[0]

		// Response struct to populate and return
		var response = struct {
			Locations db.Location     `json:"location"`
			Shifts    db.Shift        `json:"shift"`
			Vehicles  db.Vehicle      `json:"vehicle"`
			Roles     db.OperatorRole `json:"role"`
		}{}

		// Roles retrieval
		today := time.Now()

		// Retrieve day coordinates
		dayCoord := gsuite.DayCoord{}
		dayCoord.New()

		// Gsheet service
		srv := gsuite.Service{}
		srv.New(os.Getenv("SHIFT_ID"))

		// Retrieve today shift
		todayShift, err := srv.ReadDay(dayCoord, today)
		if err != nil {
			fmt.Printf("Cannot retrieve requested shift, no shift found: %v\n", err)
			return context.String(http.StatusNotFound, "Cannot retrieve requested shift, no shift found")
		}

		// Retrieve today roles
		todayRoles, err := srv.GetOperatorRoles(todayShift, name)
		if err != nil {
			fmt.Printf("cannot retrieve requested roles, operator not found: %v\n", err)
			return context.String(http.StatusNotFound, "Cannot retrieve requested roles, operator not found")
		}

		// Populate roles struct
		splitRoles := strings.Split(todayRoles, "|")
		response.Locations.Name = splitRoles[0]
		response.Shifts.Name = splitRoles[1]
		response.Vehicles.Name = splitRoles[2]
		response.Roles.Name = splitRoles[3]

		// Return today shift
		return context.JSON(http.StatusOK, response)
	}
}

func GetLoggedInOperatorShiftByDate() echo.HandlerFunc {
	return func(context echo.Context) error {
		// Read operator name fom JWT
		user := context.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		opFullName := claims["opname"]

		// Extract operator surname from claim
		name := strings.Split(opFullName.(string), " ")[0]

		// Response struct to populate and return
		var response = struct {
			Locations db.Location     `json:"location"`
			Shifts    db.Shift        `json:"shift"`
			Vehicles  db.Vehicle      `json:"vehicle"`
			Roles     db.OperatorRole `json:"role"`
		}{}

		// Roles retrieval
		date, err := time.Parse("20060102", context.Param("date"))

		// Retrieve day coordinates
		dayCoord := gsuite.DayCoord{}
		dayCoord.New()

		// Gsheet service
		srv := gsuite.Service{}
		srv.New(os.Getenv("SHIFT_ID"))

		// Retrieve today shift
		todayShift, err := srv.ReadDay(dayCoord, date)
		if err != nil {
			fmt.Printf("Cannot retrieve requested shift, no shift found: %v\n", err)
			return context.String(http.StatusNotFound, "Cannot retrieve requested shift, no shift found")
		}

		// Retrieve today roles
		todayRoles, err := srv.GetOperatorRoles(todayShift, name)
		if err != nil {
			fmt.Printf("cannot retrieve requested roles, operator not found: %v\n", err)
			return context.String(http.StatusNotFound, "Cannot retrieve requested roles, operator not found")
		}

		// Populate roles struct
		splitRoles := strings.Split(todayRoles, "|")
		response.Locations.Name = splitRoles[0]
		response.Shifts.Name = splitRoles[1]
		response.Vehicles.Name = splitRoles[2]
		response.Roles.Name = splitRoles[3]

		// Return today shift
		return context.JSON(http.StatusOK, response)
	}
}
