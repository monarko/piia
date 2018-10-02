package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

// Screening model
type Screening struct {
	ID                uuid.UUID   `json:"id" db:"id"`
	CreatedAt         time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at" db:"updated_at"`
	LeftVisualAcuity  string      `json:"left_visual_acuity" db:"left_visual_activity"`
	LeftGradingDr     string      `json:"left_grading_dr" db:"left_grading_dr"`
	LeftGradingDme    string      `json:"left_grading_dme" db:"left_grading_dme"`
	RightVisualAcuity string      `json:"right_visual_acuity" db:"right_visual_activity"`
	RightGradingDr    string      `json:"right_grading_dr" db:"right_grading_dr"`
	RightGradingDme   string      `json:"right_grading_dme" db:"right_grading_dme"`
	Referred          bool        `json:"referred" db:"referred"`
	Screener          User        `belongs_to:"user"`
	ScreenerID        uuid.UUID   `json:"screener_id" db:"screener_id"`
	Participant       Participant `belongs_to:"participant"`
	ParticipantID     uuid.UUID   `json:"participant_id" db:"participant_id"`
}

// String is not required by pop and may be deleted
func (s Screening) String() string {
	js, _ := json.Marshal(s)
	return string(js)
}

// Screenings is not required by pop and may be deleted
type Screenings []Screening

// String is not required by pop and may be deleted
func (s Screenings) String() string {
	js, _ := json.Marshal(s)
	return string(js)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (s *Screening) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: s.LeftVisualAcuity, Name: "LeftVisualAcuity"},
		&validators.StringIsPresent{Field: s.LeftGradingDr, Name: "LeftGradingDr"},
		&validators.StringIsPresent{Field: s.LeftGradingDme, Name: "LeftGradingDme"},
		&validators.StringIsPresent{Field: s.RightVisualAcuity, Name: "RightVisualAcuity"},
		&validators.StringIsPresent{Field: s.RightGradingDr, Name: "RightGradingDr"},
		&validators.StringIsPresent{Field: s.RightGradingDme, Name: "RightGradingDme"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (s *Screening) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (s *Screening) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
