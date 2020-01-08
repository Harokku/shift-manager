package db

import (
	"database/sql"
	"os"
	"testing"
	"time"
)

func TestShiftChange_GetById(t *testing.T) {
	// Service init
	dbConn, _ := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	defer dbConn.Close()
	service := Service{Db: dbConn}

	type fields struct {
		service           Service
		Id                string
		Manager           string
		Outcome           bool
		Status            string
		RequestTimestamp  time.Time
		ResponseTimestamp time.Time
		ApplicantName     string
		ApplicantDate     time.Time
		WithName          string
		WithDate          time.Time
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Test case",
			fields: fields{
				service:           service,
				Id:                "a4bf72a0-e3ea-4f5d-8d1f-1d3f456791ca",
				Manager:           "75a49c6b-5456-4ed2-9d0c-4dde2cdbe707",
				Outcome:           true,
				Status:            "pending",
				RequestTimestamp:  time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC),
				ResponseTimestamp: time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC),
				ApplicantName:     "75a49c6b-5456-4ed2-9d0c-4dde2cdbe707",
				ApplicantDate:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				WithName:          "75a49c6b-5456-4ed2-9d0c-4dde2cdbe707",
				WithDate:          time.Date(2020, 1, 2, 9, 0, 0, 0, time.UTC),
			},
			args:    args{id: "a4bf72a0-e3ea-4f5d-8d1f-1d3f456791ca"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShiftChange{
				service:           tt.fields.service,
				Id:                tt.fields.Id,
				Manager:           tt.fields.Manager,
				Outcome:           tt.fields.Outcome,
				Status:            tt.fields.Status,
				RequestTimestamp:  tt.fields.RequestTimestamp,
				ResponseTimestamp: tt.fields.ResponseTimestamp,
				ApplicantName:     tt.fields.ApplicantName,
				ApplicantDate:     tt.fields.ApplicantDate,
				WithName:          tt.fields.WithName,
				WithDate:          tt.fields.WithDate,
			}
			if err := s.GetById(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("GetById() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
