package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

// OverReading model
type OverReading struct {
	ID                uuid.UUID   `json:"id" db:"id"`
	CreatedAt         time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at" db:"updated_at"`
	LeftVisualAcuity  string      `json:"left_visual_acuity" db:"left_visual_acuity"`
	LeftGradingDr     string      `json:"left_grading_dr" db:"left_grading_dr"`
	LeftGradingDme    string      `json:"left_grading_dme" db:"left_grading_dme"`
	RightVisualAcuity string      `json:"right_visual_acuity" db:"right_visual_acuity"`
	RightGradingDr    string      `json:"right_grading_dr" db:"right_grading_dr"`
	RightGradingDme   string      `json:"right_grading_dme" db:"right_grading_dme"`
	Referred          bool        `json:"referred" db:"referred"`
	OverReader        User        `belongs_to:"user"`
	OverReaderID      uuid.UUID   `json:"over_reader_id" db:"over_reader_id"`
	Participant       Participant `belongs_to:"participant"`
	ParticipantID     uuid.UUID   `json:"participant_id" db:"participant_id"`
}

// String is not required by pop and may be deleted
func (o OverReading) String() string {
	jo, _ := json.Marshal(o)
	return string(jo)
}

// OverReadings is not required by pop and may be deleted
type OverReadings []OverReading

// String is not required by pop and may be deleted
func (o OverReadings) String() string {
	jo, _ := json.Marshal(o)
	return string(jo)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (o *OverReading) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: o.LeftVisualAcuity, Name: "LeftVisualAcuity"},
		&validators.StringIsPresent{Field: o.LeftGradingDr, Name: "LeftGradingDr"},
		&validators.StringIsPresent{Field: o.LeftGradingDme, Name: "LeftGradingDme"},
		&validators.StringIsPresent{Field: o.RightVisualAcuity, Name: "RightVisualAcuity"},
		&validators.StringIsPresent{Field: o.RightGradingDr, Name: "RightGradingDr"},
		&validators.StringIsPresent{Field: o.RightGradingDme, Name: "RightGradingDme"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (o *OverReading) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (o *OverReading) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
