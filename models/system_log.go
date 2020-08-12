package models

import (
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// SystemLog model
type SystemLog struct {
	ID           uuid.UUID `json:"id" db:"id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	Action       string    `json:"action" db:"action"`
	Activity     string    `json:"activity" db:"activity"`
	Error        bool      `json:"error" db:"error"`
	ErrorMessage string    `json:"error_message" db:"error_message"`
	ClientIP     string    `json:"client_ip" db:"client_ip"`
	User         User      `belongs_to:"user" json:"user"`
	UserID       uuid.UUID `json:"-" db:"user_id"`
	ResourceID   string    `json:"resource_id" db:"resource_id"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
}

// SystemLogs is not required by pop and may be deleted
type SystemLogs []SystemLog

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (s *SystemLog) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: s.Action, Name: "Action"},
		&validators.StringIsPresent{Field: s.ClientIP, Name: "ClientIP"},
	), nil
}
