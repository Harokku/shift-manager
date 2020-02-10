package api

import (
	"database/sql"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"net/http"
	"os"
	"shift-manager/db"
	"shift-manager/gsuite"
	"time"
)

type change struct {
	FirstDate  time.Time `json:"first_date"`
	FirstName  string    `json:"first_name"`
	SecondDate time.Time `json:"second_date"`
	SecondName string    `json:"second_name"`
}

// PutChange actually modify gsheet shift table switching passed operators
//
// Request body:
// {
//		first_date: Requester date
//		first_name: Requester operator name
//		second_date: Requested date
//		second_name: Requested operator name
// }
func PutChange() echo.HandlerFunc {
	return func(context echo.Context) error {
		var err error
		sheetService := gsuite.Service{}
		err = sheetService.New(os.Getenv("SHIFT_ID"))
		if err != nil {
			fmt.Printf("Error creating gSheet service: %v\n", err)
			return context.String(http.StatusInternalServerError, fmt.Sprintf("Error creating gSheet service: %v\n", err))
		}

		var c change

		// Bind request body to change struct
		if err := context.Bind(&c); err != nil {
			fmt.Printf("Error binding request body: %v\n", err)
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error binding request body: %v\n", err))
		}

		// Create an instance of gsuite.ShiftToSwitch and populate fields from request body
		sc := gsuite.ShiftsToSwitch{}
		err = sc.New(sheetService)
		if err != nil {
			fmt.Printf("Error creating shift change service: %v\n", err)
			return context.String(http.StatusInternalServerError, fmt.Sprintf("Error creating shift change service: %v\n", err))
		}

		sc.FirstDate = c.FirstDate
		sc.FirstName = c.FirstName
		sc.SecondDate = c.SecondDate
		sc.SecondName = c.SecondName

		// Call service to actually modify gsheet
		err = sc.SwitchShifts()
		if err != nil {
			fmt.Printf("Error switching shifts: %v,\n", err)
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error switching shifts: %v,\n", err))
		}
		return context.String(http.StatusOK, "Shift correctly modified")
	}
}

// RequestChange create a new shift change request to DB. Will be posted to gsheet after is been managed
//
// Request body:
// {
//		applicant_date: Requester date
//		with_date: Requested date
//		with_name: Requested operator name
// }
func RequestChange(s *db.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		var (
			err         error
			requester   db.User        // Requester user data (username and ID)
			shiftChange db.ShiftChange // Shift change service
		)

		// Read user from JWT and extract claims
		user := context.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		username := claims["username"].(string)

		// Create service and get logged in user's DB ID
		requester.New(*s)
		err = requester.GetUser(username)
		if err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("error retrieving user's ID: %v\n", err))
		}

		// Populate shiftChange with request body data
		shiftChange.New(*s)
		if err = context.Bind(&shiftChange); err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("error binding request body: %v\n", err))
		}
		shiftChange.ApplicantName = requester.Id
		err = shiftChange.NewRequest()
		if err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("error creating shift request: %v\n", err))
		}

		return context.String(http.StatusOK, "Shift change request correctly submitted")
	}
}

// ManageChangeRequest set change request as accepted or refused and update db and gsheet accordingly
//
// Will read actual manager name from JWT and set timestamp and outcome automatically.
//
// Unused field will be discarded
//
// Request body:
// {
//		id: change request id
//		status: one of "rejected" or "accepted"
// }
// TODO: implement func
func ManageChangeRequest(s *db.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		type param struct {
			Id     string `json:"id"`
			Status string `json:"status"`
		}

		type manager struct {
			id   string
			name string
		}

		var (
			err            error
			p              param
			sheetService   gsuite.Service
			sc             gsuite.ShiftsToSwitch
			statusToChange db.ShiftChange
			m              manager
		)

		// Retrieve manager name from DB based upon logged in user
		sqlGetManagerNameFromUsername := `
			select 
			       o."user",
			       o.surname
			from users u
					 inner join operators o on u.id = o."user"
			where u.username = $1
			`
		// Read user from JWT, extract claims and populate manager name
		user := context.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		username := claims["username"].(string)
		row := s.Db.QueryRow(sqlGetManagerNameFromUsername, username)
		switch err = row.Scan(&m.id, &m.name); err {
		case sql.ErrNoRows:
			fmt.Printf("No manager name found: %v\n", err)
			return context.String(http.StatusNotFound, fmt.Sprintf("No manager name found: %v\n", err))
		case nil:
		default:
			fmt.Printf("Error retrieving manager")
			return context.String(http.StatusInternalServerError, fmt.Sprintf("Error retrieving manager name: %v\n", err))
		}

		// Bind request body to param struct for further user
		if err = context.Bind(&p); err != nil {
			fmt.Printf("Error binding request body: %v\n", err)
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error binding request body: %v\n", err))
		}

		// create new shiftChange service
		statusToChange.New(*s)
		err = statusToChange.GetById(p.Id)
		if err != nil {
			fmt.Printf("Error retrieving selected change request: %v\n", err)
			return context.String(http.StatusNotFound, fmt.Sprintf("Error retrieving selected change request: %v\n", err))
		}
		// Set shift change managerId and status from request data
		statusToChange.Manager = m.id
		statusToChange.Status = p.Status

		// -------------
		// Prepare gsheet service and modify spreadsheet, then, if succesful update DB
		// -------------

		// create new gsheet service
		err = sheetService.New(os.Getenv("SHIFT_ID"))
		if err != nil {
			fmt.Printf("Error creating gSheet service: %v\n", err)
			return context.String(http.StatusInternalServerError, fmt.Sprintf("Error creating gSheet service: %v\n", err))
		}
		err = sc.New(sheetService)
		if err != nil {
			fmt.Printf("Error creating gSheet service: %v\n", err)
			return context.String(http.StatusInternalServerError, fmt.Sprintf("Error creating gSheet service: %v\n", err))
		}

		sqlSheetChange := `
			select c.applicant_date,
				   a.surname as applicant_name,
				   c.with_date,
				   w.surname as with_name
			from shift_change c
					 inner join operators a on c.applicant_name = a."user"
					 inner join operators w on c.with_name = w."user"
			where c.id = $1
			`

		// populate sc with with applicant and with data from passed change request id
		row = s.Db.QueryRow(sqlSheetChange, p.Id)
		switch err = row.Scan(&sc.FirstDate, &sc.FirstName, &sc.SecondDate, &sc.SecondName); err {
		case sql.ErrNoRows:
			fmt.Printf("No requester found: %v\n", err)
			return context.String(http.StatusNotFound, fmt.Sprintf("No requester found: %v\n", err))
		case nil:
		default:
			fmt.Printf("Error retrieving requester")
			return context.String(http.StatusInternalServerError, fmt.Sprintf("Error retrieving requester name: %v\n", err))
		}

		// call service to actually modify gsheet
		err = sc.SwitchShifts()
		if err != nil {
			fmt.Printf("Error switching shifts: %v\n", err)
			return context.String(http.StatusBadRequest, fmt.Sprintf("Error switching shifts: %v\n", err))
		}

		// call db service to update status
		err = statusToChange.ChangeStatus()
		if err != nil {
			fmt.Printf("Error updating change request: %v\n", err)
			return context.String(http.StatusInternalServerError, fmt.Sprintf("Error updating change request: %v\n", err))
		}

		return context.String(http.StatusOK, "change request managed")
	}
}

// GetAllChanges return all changes
func GetAllChanges(s *db.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		var (
			err          error
			shiftChange  db.ShiftChange
			shiftChanges []db.ShiftChange
		)

		shiftChange.New(*s)
		err = shiftChange.GetAll(&shiftChanges)
		if err != nil {
			return context.String(http.StatusNotFound, fmt.Sprintf("no shift change found: %v\n", err))
		}

		return context.JSON(http.StatusOK, shiftChanges)
	}
}

func GetAllChangesForUser(s *db.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		var (
			err          error
			requester    db.User
			shiftChange  db.ShiftChange
			shiftChanges []db.ShiftChange
		)

		// Read user from JWT and extract claims
		user := context.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		username := claims["username"].(string)

		// Create service and get logged in user's DB ID
		requester.New(*s)
		err = requester.GetUser(username)
		if err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("error retrieving user's ID: %v\n", err))
		}

		shiftChange.New(*s)
		shiftChange.ApplicantName = requester.Id
		err = shiftChange.GetAllByApplicant(&shiftChanges)
		if err != nil {
			return context.String(http.StatusBadRequest, fmt.Sprintf("error retrieving user's shift changes: %v\n", err))
		}

		return context.JSON(http.StatusOK, shiftChanges)
	}
}
