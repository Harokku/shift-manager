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
					WHERE id = $1
					ORDER BY request_timestamp DESC`
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
