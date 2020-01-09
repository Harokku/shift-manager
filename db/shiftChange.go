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
	Manager           string    `json:"manager"`
	Outcome           bool      `json:"outcome"`
	Status            string    `json:"status"`
	RequestTimestamp  time.Time `json:"request_timestamp"`
	ResponseTimestamp time.Time `json:"response_timestamp"`
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
	sqlStatement := `SELECT id,
						   manager_name,
						   outcome,
						   status,
						   request_timestamp,
						   response_timestamp,
						   applicant_name,
						   applicant_date,
						   with_name,
						   with_date
					FROM shift_change
					WHERE id = $1`

	row := s.service.Db.QueryRow(sqlStatement, id)
	switch err := row.Scan(&s.Id, &s.Manager, &s.Outcome, &s.Status, &s.RequestTimestamp, &s.ResponseTimestamp, &s.ApplicantName, &s.ApplicantDate, &s.WithName, &s.WithDate); err {
	case sql.ErrNoRows:
		return errors.New("no row where retrieved")
	case nil:
		return nil
	default:
		return errors.New(fmt.Sprintf("error retrieving shift change from database: %v\n", err))
	}
}

// GetAdd retrieve all shift changes from db, ordered from newest to older
//
// dest []ShiftChange: You must pass an array pointer to ShiftChange who will be populated with retrieved content
func (s *ShiftChange) GetAll(dest *[]ShiftChange) error {
	sqlStatement := `SELECT id,
						   manager_name,
						   outcome,
						   status,
						   request_timestamp,
						   response_timestamp,
						   applicant_name,
						   applicant_date,
						   with_name,
						   with_date
					FROM shift_change
					ORDER BY request_timestamp DESC`

	rows, err := s.service.Db.Query(sqlStatement)
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
