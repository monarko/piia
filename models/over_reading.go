package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/monarko/piia/helpers/types"
)

// OverReading model
type OverReading struct {
	ID            uuid.UUID               `json:"id" db:"id"`
	CreatedAt     time.Time               `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at" db:"updated_at"`
	Eyes          types.EyeOverReading    `json:"eyes" db:"eye_assessment"`
	Referral      types.ReferralScreening `json:"referral" db:"referral"`
	OverReader    User                    `belongs_to:"user" json:"over_reader"`
	OverReaderID  uuid.UUID               `json:"-" db:"over_reader_id"`
	Participant   Participant             `belongs_to:"participant" json:"participant"`
	ParticipantID uuid.UUID               `json:"-" db:"participant_id"`
	Screening     Screening               `belongs_to:"screening" json:"screening"`
	ScreeningID   uuid.UUID               `json:"-" db:"screening_id"`
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
	// &validators.StringIsPresent{Field: o.Eyes.LeftEye.DRGrading, Name: "LeftGradingDr"},
	// &validators.StringIsPresent{Field: o.Eyes.LeftEye.DMEAssessment, Name: "LeftGradingDme"},
	// &validators.StringIsPresent{Field: o.Eyes.RightEye.DRGrading, Name: "RightGradingDr"},
	// &validators.StringIsPresent{Field: o.Eyes.RightEye.DMEAssessment, Name: "RightGradingDme"},
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
