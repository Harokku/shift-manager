package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type ShiftChange struct {
	service           Service
	Id                string    `json:"id"`
	Manager           string    `json:"manager,omitempty"`
	Outcome           bool      `json:"outcome"`
	Status            string    `json:"status"`
	RequestTimestamp  time.Time `json:"request_timestamp"`
	ResponseTimestamp time.Time `json:"response_timestamp,omitempty"`
	ApplicantName     string    `json:"applicant_name"`
	ApplicantDate     time.Time `json:"applicant_date"`
	WithName          string    `json:"with_name"`
	WithDate          time.Time `json:"with_date"`
}

func (s *ShiftChange) New(service Service) {
	s.service = service
}

// GetById retrieve shift change request from db, filtered by passed ID, return error if not found
//
// s ShiftChange: struct populated if successfully retrieved shift change
func (s *ShiftChange) GetById(id string) error {
	nullTime := time.Time{}
	sqlStatement := `SELECT id,
						   COALESCE(CAST(manager_name as varchar),'') as manager_name,
						   outcome,
						   status,
						   request_timestamp,
						   COALESCE(response_timestamp, $2) as response_timestamp,
						   applicant_name,
						   applicant_date,
						   with_name,
						   with_date
					FROM shift_change
					WHERE id = $1`

	row := s.service.Db.QueryRow(sqlStatement, id, nullTime)
	switch err := row.Scan(&s.Id, &s.Manager, &s.Outcome, &s.Status, &s.RequestTimestamp, &s.ResponseTimestamp, &s.ApplicantName, &s.ApplicantDate, &s.WithName, &s.WithDate); err {
	case sql.ErrNoRows:
		return errors.New("no row where retrieved")
	case nil:
		return nil
	default:
		return errors.New(fmt.Sprintf("error retrieving shift change from database: %v\n", err))
	}
}

// GetAll retrieve all shift changes from db, ordered from newest to older
//
// dest []ShiftChange: You must pass an array pointer to ShiftChange who will be populated with retrieved content
func (s *ShiftChange) GetAll(dest *[]ShiftChange) error {
	nullTime := time.Time{}
	sqlStatement := `SELECT s.id,
						   COALESCE(CAST(s.manager_name as varchar), '') as manager_name,
						   s.outcome,
						   s.status,
						   s.request_timestamp,
						   COALESCE(s.response_timestamp, $1) as response_timestamp,
						   CONCAT(a.surname, ' ', a.name)                as applicant_surname,
						   s.applicant_date,
						   CONCAT(w.surname, ' ', w.name)                 as with_surname,
						   s.with_date
					FROM shift_change s
						INNER JOIN operators a on s.applicant_name = a."user"
						INNER JOIN operators w on s.with_name = w."user"
					ORDER BY s.applicant_date DESC`

	rows, err := s.service.Db.Query(sqlStatement, nullTime)
	if err != nil {
		return errors.New(fmt.Sprintf("error retrieving shifts change: %v\n", err))
	}
	defer rows.Close()

	for rows.Next() {
		var shiftChange ShiftChange
		err = rows.Scan(&shiftChange.Id, &shiftChange.Manager, &shiftChange.Outcome, &shiftChange.Status, &shiftChange.RequestTimestamp, &shiftChange.ResponseTimestamp, &shiftChange.ApplicantName, &shiftChange.ApplicantDate, &shiftChange.WithName, &shiftChange.WithDate)
		if err != nil {
			return errors.New(fmt.Sprintf("error scanning row: %v\n", err))
		}
		*dest = append(*dest, shiftChange)
	}
	err = rows.Err()
	if err != nil {
		return errors.New(fmt.Sprintf("error appending rows to result: %v\n", err))
	}

	return nil
}

// GetAllByApplicant retrieve all shift changes requests from DB based on applicant UUID
//
// dest []ShiftChange: You must pass an array pointer to ShiftChange who will be populated with retrieved content
//
// set applicant_name UUID before call
func (s *ShiftChange) GetAllByApplicant(dest *[]ShiftChange) error {
	nulltime := time.Time{}
	sqlStatement := `SELECT s.id,
						   COALESCE(CAST(s.manager_name as varchar), '') as manager_name,
						   s.outcome,
						   s.status,
						   s.request_timestamp,
						   COALESCE(s.response_timestamp, $1) as response_timestamp,
						   a.surname as applicant_name,
						   s.applicant_date,
						   w.surname as with_name,
						   s.with_date
					FROM shift_change s
						INNER JOIN operators a ON s.applicant_name = a."user"
         				INNER JOIN operators w ON s.with_name = w."user"
					WHERE applicant_name = $2
					ORDER BY request_timestamp DESC
`

	rows, err := s.service.Db.Query(sqlStatement, nulltime, s.ApplicantName)
	if err != nil {
		return errors.New(fmt.Sprintf("error retrieving shift changes: %v\n", err))
	}
	defer rows.Close()

	for rows.Next() {
		var shiftChange ShiftChange
		err = rows.Scan(&shiftChange.Id, &shiftChange.Manager, &shiftChange.Outcome, &shiftChange.Status, &shiftChange.RequestTimestamp, &shiftChange.ResponseTimestamp, &shiftChange.ApplicantName, &shiftChange.ApplicantDate, &shiftChange.WithName, &shiftChange.WithDate)
		if err != nil {
			return errors.New(fmt.Sprintf("error scanning row: %v\n", err))
		}
		*dest = append(*dest, shiftChange)
	}
	err = rows.Err()
	if err != nil {
		return errors.New(fmt.Sprintf("error appending rows to result: %v\n", err))
	}
	return nil
}

// NewRequest create a new shift request, setting initial status
//
// Populate required field before invoke:
// ApplicantName, ApplicantDate, WithName, WithDate
//
// Applicant is the operator who ask for change
//
// With is the operator to change with
func (s ShiftChange) NewRequest() error {
	sqlStatement := `
					INSERT INTO shift_change (applicant_name, applicant_date, with_name, with_date)
					VALUES ($1,$2,$3,$4)
`
	_, err := s.service.Db.Exec(sqlStatement, s.ApplicantName, s.ApplicantDate, s.WithName, s.WithDate)
	if err != nil {
		return errors.New(fmt.Sprintf("error creating new shift change request: %v\n", err))
	}
	return nil
}

// ChangeStatus update change request status
//
// set required fields in struct before invoking, non required fields will be discarded:
// ID, Manager, Status
//
// If successful set outcome to true (standing shift request has been evaded and require no further attention) and response timestamp
func (s *ShiftChange) ChangeStatus() error {
	timestamp := time.Now()
	sqlStatement := `
					UPDATE shift_change
					SET manager_name=$2, 
					    outcome=true,
					    status=$3,
					    response_timestamp=$4
					WHERE id=$1
`
	_, err := s.service.Db.Exec(sqlStatement, s.Id, s.Manager, s.Status, timestamp)
	if err != nil {
		return errors.New(fmt.Sprintf("error updating status: %v\n", err))
	}
	s.Outcome = true
	return nil
}
