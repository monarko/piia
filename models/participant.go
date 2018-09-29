package models

import (
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

// Participant object
type Participant struct {
	ID            uuid.UUID `json:"id" db:"id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	ParticipantID string    `json:"participant_id" db:"participant_id"`
	Name          string    `json:"name" db:"name"`
	Gender        string    `json:"gender" db:"gender"`
	DOB           time.Time `json:"dob" db:"dob"`
	ContactNumber string    `json:"contact_number" db:"contact_number"`
	IsEligible    bool      `json:"is_eligible" db:"is_eligible"`
	Consented     bool      `json:"consented" db:"consented"`
	AuthorID      uuid.UUID `json:"author_id" db:"author_id"`
}

// Participants holds the list of Participant
type Participants []Participant

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (p *Participant) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: p.Name, Name: "Name"},
		&validators.StringIsPresent{Field: p.Gender, Name: "Gender"},
		&validators.TimeIsPresent{Field: p.DOB, Name: "Date of birth"},
	), nil
}
